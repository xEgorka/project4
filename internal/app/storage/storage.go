// Package storage implements Storage interface.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/logger"
	"github.com/xEgorka/project4/internal/app/models"
)

// Storage describes methods required to implement Storage.
type Storage interface {
	Add(ctx context.Context, d models.Song) (models.Song, error)
	Update(ctx context.Context, id string, data models.RequestUpdateSong) error
	Delete(ctx context.Context, id string) error
	GetText(ctx context.Context, id string, page, size int) (models.ResponseGetSongText, error)
	GetSongs(ctx context.Context, d models.Song, page, size int) (models.ResponseGetSongs, error)
	Ping() error
	Close() error
}

// Open initializes Storage.
func Open(ctx context.Context, cfg *config.Config) (Storage, error) {
	logger.Log.Info("opening database...", zap.String("conninfo", cfg.DBURI))
	conn, err := sql.Open(cfg.DBDriver, cfg.DBURI)
	if err != nil {
		return nil, err
	}
	if e := conn.Ping(); e != nil {
		return nil, e
	}
	return open(cfg, conn)
}

func open(cfg *config.Config, conn *sql.DB) (Storage, error) {
	db := new(cfg, conn)
	if err := bootstrap(conn); err != nil {
		logger.Log.Error("failed bootstrap", zap.Error(err))
		return nil, err
	}
	return db, nil
}

type db struct {
	conn *sql.DB
	cfg  *config.Config
}

func new(config *config.Config, conn *sql.DB) *db { return &db{cfg: config, conn: conn} }

func bootstrap(conn *sql.DB) error {
	logger.Log.Info("running migrations...")
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		logger.Log.Error("failed init driver", zap.Error(err))
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logger.Log.Error("failed init migrate", zap.Error(err))
		return err
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Log.Info("running migrations... no changes")
		} else {
			logger.Log.Error("running migrations... failed", zap.Error(err))
			return err
		}
	} else {
		logger.Log.Info("running migrations... changes applied")
	}
	return nil
}

// Ping checks database connection.
func (s *db) Ping() error { return s.conn.Ping() }

// Close closes database connection.
func (s *db) Close() error { return s.conn.Close() }

const (
	queryInsertSong = `
insert into songs (id, "group", song, release_date, text, link) values ($1, $2, $3, $4, $5, $6)
`
	querySelectSong = `select id from songs where "group"=$1 and song=$2 and deleted=False`
)

// ErrUniqueViolation indicates song unique constraint violation.
var ErrUniqueViolation = errors.New(`ERROR: duplicate key value violates unique constraint "songs_idx" (SQLSTATE 23505)`)

// Add creates song in database.
func (s *db) Add(ctx context.Context, song models.Song) (models.Song, error) {
	id := uuid.New().String()
	_, err := s.conn.ExecContext(ctx, queryInsertSong, id,
		song.Group, song.Song, song.ReleaseDate, song.Text, song.Link)
	if err != nil && err.Error() == ErrUniqueViolation.Error() {
		row := s.conn.QueryRowContext(ctx, querySelectSong, song.Group, song.Song)
		var id string
		if err := row.Scan(&id); err != nil {
			return models.Song{}, err // ErrNoRows if song deleted
		}
		return models.Song{}, ErrUniqueViolation
	} else if err != nil {
		return models.Song{}, err
	}
	song.ID = id
	return song, nil
}

const queryUpdateSong = `update songs set release_date=$2, text=$3, link=$4 where id=$1 and deleted=False`

// ErrNotAffected indicates no row affected as a result of the query.
var ErrNotAffected = errors.New(`not affected`)

// Update updates song in library.
func (s *db) Update(ctx context.Context, id string, d models.RequestUpdateSong) error {
	res, err := s.conn.ExecContext(ctx, queryUpdateSong, id, d.ReleaseDate, d.Text, d.Link)
	if err != nil {
		return err
	}
	row, e := res.RowsAffected()
	if e != nil {
		return e
	}
	if row < 1 {
		return ErrNotAffected
	}
	return nil
}

const queryDeleteSong = `update songs set deleted=True where id=$1 and deleted=False`

// Delete soft deletes song from library.
func (s *db) Delete(ctx context.Context, id string) error {
	res, err := s.conn.ExecContext(ctx, queryDeleteSong, id)
	if err != nil {
		return err
	}
	row, e := res.RowsAffected()
	if e != nil {
		return e
	}
	if row < 1 {
		return ErrNotAffected
	}
	return nil
}

const querySelectSongText = `select "group", song, text from songs where id=$1 and deleted=False`

// GetText returns song text.
func (s *db) GetText(ctx context.Context, id string,
	page, size int) (models.ResponseGetSongText, error) {
	row := s.conn.QueryRowContext(ctx, querySelectSongText, id)
	var group, song, text string
	if err := row.Scan(&group, &song, &text); err != nil {
		return models.ResponseGetSongText{}, err
	}

	v := strings.Split(text, "\n\n")
	total := len(v)
	d := models.ResponseGetSongText{
		ID:    id,
		Group: group,
		Song:  song,
		Total: total,
		Page:  page,
		Size:  size,
	}
	beg := (page - 1) * size
	if beg > total {
		return d, nil
	}
	end := beg + size
	if end > total {
		end = total
	}
	d.Verses = v[beg:end]
	return d, nil
}

// GetSongs returns filtered and paginated songs.
func (s *db) GetSongs(ctx context.Context, d models.Song,
	page, size int) (models.ResponseGetSongs, error) {
	q := `select id, "group", song, release_date, text, link from songs where deleted=False`
	args := make([]interface{}, 0)
	num := 1
	if d.ID != `` {
		q += fmt.Sprintf(` and id=$%d`, num)
		args = append(args, d.ID)
		num += 1
	}
	if d.Group != `` {
		q += fmt.Sprintf(` and "group"=$%d`, num)
		args = append(args, d.Group)
		num += 1
	}
	if d.Song != `` {
		q += fmt.Sprintf(` and song=$%d`, num)
		args = append(args, d.Song)
		num += 1
	}
	if !d.ReleaseDate.IsZero() {
		q += fmt.Sprintf(` and release_date=$%d`, num)
		args = append(args, d.ReleaseDate)
		num += 1
	}
	if d.Text != `` {
		q += fmt.Sprintf(` and text like $%d`, num)
		args = append(args, "%"+d.Text+"%")
		num += 1
	}
	if d.Link != `` {
		q += fmt.Sprintf(` and link=$%d`, num)
		args = append(args, d.Link)
		num += 1
	}
	q += fmt.Sprintf(` offset $%d limit $%d`, num, num+1)
	args = append(args, (page-1)*size)
	args = append(args, size)

	logger.Log.Debug("executing", zap.String("query", q))
	rows, err := s.conn.QueryContext(ctx, q, args...)
	if err != nil {
		return models.ResponseGetSongs{}, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			logger.Log.Error("failed close rows", zap.Error(err))
		}
	}()

	var dd []models.Song
	for rows.Next() {
		var id, group, song, text, releaseDateStr, link string
		if err = rows.Scan(&id, &group, &song,
			&releaseDateStr, &text, &link); err != nil {
			return models.ResponseGetSongs{}, err
		}
		releaseDate, e := time.Parse(time.RFC3339, releaseDateStr)
		if e != nil {
			return models.ResponseGetSongs{}, e
		}
		dd = append(dd, models.Song{
			ID:          id,
			Group:       group,
			Song:        song,
			ReleaseDate: releaseDate,
			Text:        text,
			Link:        link})
	}
	if err = rows.Err(); err != nil {
		return models.ResponseGetSongs{}, err
	}

	return models.ResponseGetSongs{Songs: dd, Page: page, Size: size}, nil
}
