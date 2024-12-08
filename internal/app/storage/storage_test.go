package storage

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/models"
)

func TestOpen(t *testing.T) {
	type args struct {
		ctx context.Context
		cfg *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    Storage
		wantErr bool
	}{
		{name: "negative test #1", args: args{ctx: context.Background(), cfg: &config.Config{DBDriver: "bad"}}, wantErr: true},
		{name: "negative test #2", args: args{ctx: context.Background(), cfg: &config.Config{DBDriver: "pgx"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Open(tt.args.ctx, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Open() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_open(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()
	type args struct {
		cfg  *config.Config
		conn *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    Storage
		wantErr bool
	}{
		{
			name:    "positive test #1",
			args:    args{cfg: &config.Config{}, conn: conn},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := open(tt.args.cfg, tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_db_Ping(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()
	tests := []struct {
		name    string
		wantErr bool
	}{{name: "positive test #1", wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &db{conn: conn}
			mock.ExpectPing()
			if err := s.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("db.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_Close(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()
	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "positive test #1", wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectClose()
			s := &db{conn: conn}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("db.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_db_Add(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()

	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	type args struct {
		ctx  context.Context
		song models.Song
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Song
		wantErr bool
	}{
		{
			name:    "negative test #1",
			args:    args{ctx: context.Background(), song: models.Song{}},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
		{
			name:    "positive test #1",
			args:    args{ctx: context.Background(), song: models.Song{}},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := db{conn: tt.fields.conn, cfg: tt.fields.cfg}
			if tt.name == "negative test #1" {
				mock.ExpectExec(regexp.QuoteMeta(queryInsertSong)).WillReturnError(errors.New("test"))
			}
			if tt.name == "positive test #1" {
				res := sqlmock.NewResult(1, 1)
				mock.ExpectExec(regexp.QuoteMeta(queryInsertSong)).WillReturnResult(res)
			}
			_, err := s.Add(tt.args.ctx, tt.args.song)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_db_Update(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()

	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	type args struct {
		ctx context.Context
		id  string
		req models.RequestUpdateSong
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Song
		wantErr bool
	}{

		{
			name:    "positive test #1",
			args:    args{ctx: context.Background(), req: models.RequestUpdateSong{}},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name:    "negative test #1",
			args:    args{ctx: context.Background(), req: models.RequestUpdateSong{}},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
		{
			name:    "negative test #2",
			args:    args{ctx: context.Background(), req: models.RequestUpdateSong{}},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := db{
				conn: tt.fields.conn,
				cfg:  tt.fields.cfg,
			}
			if tt.name == "positive test #1" {
				res := sqlmock.NewResult(1, 1)
				mock.ExpectExec(regexp.QuoteMeta(queryUpdateSong)).WillReturnResult(res)
			}
			if tt.name == "negative test #1" {
				mock.ExpectExec(regexp.QuoteMeta(queryUpdateSong)).WillReturnError(errors.New("test"))
			}
			if tt.name == "negative test #2" {
				res := sqlmock.NewResult(1, 0)
				mock.ExpectExec(regexp.QuoteMeta(queryUpdateSong)).WillReturnResult(res)
			}
			err := s.Update(tt.args.ctx, tt.args.id, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_db_Delete(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()

	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Song
		wantErr bool
	}{

		{
			name:    "positive test #1",
			args:    args{ctx: context.Background()},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name:    "negative test #1",
			args:    args{ctx: context.Background()},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
		{
			name:    "negative test #2",
			args:    args{ctx: context.Background()},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := db{
				conn: tt.fields.conn,
				cfg:  tt.fields.cfg,
			}
			if tt.name == "positive test #1" {
				res := sqlmock.NewResult(1, 1)
				mock.ExpectExec(regexp.QuoteMeta(queryDeleteSong)).WillReturnResult(res)
			}
			if tt.name == "negative test #1" {
				mock.ExpectExec(regexp.QuoteMeta(queryDeleteSong)).WillReturnError(errors.New("test"))
			}
			if tt.name == "negative test #2" {
				res := sqlmock.NewResult(1, 0)
				mock.ExpectExec(regexp.QuoteMeta(queryDeleteSong)).WillReturnResult(res)
			}
			err := s.Delete(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_db_GetText(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()

	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	type args struct {
		ctx  context.Context
		id   string
		page int
		size int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Song
		wantErr bool
	}{
		{
			name:    "positive test #1",
			args:    args{ctx: context.Background()},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name:    "positive test #2",
			args:    args{ctx: context.Background(), page: 99, size: 1},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name:    "positive test #3",
			args:    args{ctx: context.Background(), page: 1, size: 99},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name:    "negative test #1",
			args:    args{ctx: context.Background()},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := db{conn: tt.fields.conn, cfg: tt.fields.cfg}
			mockRows := sqlmock.NewRows(
				[]string{"group", "song", "text"}).
				AddRow("Muse", "Supermassive Black Hole", "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight")
			if tt.name == "negative test #1" {
				mock.ExpectQuery(regexp.QuoteMeta(querySelectSongText)).
					WithArgs(tt.args.id).WillReturnError(errors.New("test"))
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(querySelectSongText)).
					WithArgs(tt.args.id).WillReturnRows(mockRows)
			}

			_, err := s.GetText(tt.args.ctx, tt.args.id, tt.args.page, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.GetText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_db_GetSongs(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer conn.Close()

	type fields struct {
		conn *sql.DB
		cfg  *config.Config
	}
	type args struct {
		ctx  context.Context
		song models.Song
		page int
		size int
	}
	releaseDateStr := "16.07.2006"
	releaseDate, ee := time.Parse("02.01.2006", releaseDateStr)
	if ee != nil {
		panic(ee)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Song
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID: "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "positive test #2",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:    "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group: "Muse",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "positive test #3",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:    "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group: "Muse",
					Song:  "Super Muse Black Hole",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "positive test #4",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:          "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group:       "Muse",
					Song:        "Super Muse Black Hole",
					ReleaseDate: releaseDate,
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "positive test #5",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:          "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group:       "Muse",
					Song:        "Super Muse Black Hole",
					ReleaseDate: releaseDate,
					Text:        "baby",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "positive test #6",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:          "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group:       "Muse",
					Song:        "Super Muse Black Hole",
					ReleaseDate: releaseDate,
					Text:        "baby",
					Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: false,
		},
		{
			name: "negative test #1",
			args: args{ctx: context.Background(),
				song: models.Song{
					ID:          "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
					Group:       "Muse",
					Song:        "Super Muse Black Hole",
					ReleaseDate: releaseDate,
					Text:        "baby",
					Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
				},
				page: 1,
				size: 10,
			},
			fields:  fields{conn: conn, cfg: &config.Config{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := db{conn: tt.fields.conn, cfg: tt.fields.cfg}
			q := `select id, "group", song, release_date, text, link from songs where deleted=False`
			mockRows := sqlmock.NewRows(
				[]string{"id", "group", "song", "release_date", "text", "link"}).
				AddRow("0824f9fb-7397-4f19-95d5-f9ce8bec75de", "Muse", "Supermassive Black Hole", "2006-07-16T00:00:00Z", "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight", "https://www.youtube.com/watch?v=Xsp3_a-PMTw")
			if tt.name == "positive test #1" {
				query := q + " and id=$1"
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.page-1, tt.args.size).WillReturnRows(mockRows)
			}
			if tt.name == "positive test #2" {
				query := q + ` and id=$1 and "group"=$2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.page-1, tt.args.size).
					WillReturnRows(mockRows)
			}
			if tt.name == "positive test #3" {
				query := q + ` and id=$1 and "group"=$2 and song=$3`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.song.Song,
						tt.args.page-1, tt.args.size).
					WillReturnRows(mockRows)
			}
			if tt.name == "positive test #4" {
				query := q + ` and id=$1 and "group"=$2 and song=$3 and release_date=$4`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.song.Song,
						tt.args.song.ReleaseDate, tt.args.page-1, tt.args.size).
					WillReturnRows(mockRows)
			}
			if tt.name == "positive test #5" {
				query := q + ` and id=$1 and "group"=$2 and song=$3 and release_date=$4 and text like $5`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.song.Song,
						tt.args.song.ReleaseDate, "%"+tt.args.song.Text+"%", tt.args.page-1, tt.args.size).
					WillReturnRows(mockRows)
			}
			if tt.name == "positive test #6" {
				query := q + ` and id=$1 and "group"=$2 and song=$3 and release_date=$4 and text like $5 and link=$6`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.song.Song,
						tt.args.song.ReleaseDate, "%"+tt.args.song.Text+"%", tt.args.song.Link,
						tt.args.page-1, tt.args.size).WillReturnRows(mockRows)
			}
			if tt.name == "negative test #1" {
				query := q + ` and id=$1 and "group"=$2 and song=$3 and release_date=$4 and text like $5 and link=$6`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.args.song.ID, tt.args.song.Group, tt.args.song.Song,
						tt.args.song.ReleaseDate, "%"+tt.args.song.Text+"%", tt.args.song.Link,
						tt.args.page-1, tt.args.size).WillReturnError(errors.New("test"))
			}

			_, err := s.GetSongs(tt.args.ctx, tt.args.song, tt.args.page, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.GetSongs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
