// Package service implements system facade.
package service

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/logger"
	"github.com/xEgorka/project4/internal/app/models"
	"github.com/xEgorka/project4/internal/app/requests"
	"github.com/xEgorka/project4/internal/app/storage"
)

// Service provides business logic.
type Service struct {
	cfg *config.Config
	s   storage.Storage
	r   *requests.HTTP
}

// New creates Service.
func New(config *config.Config, store storage.Storage, requests *requests.HTTP) *Service {
	return &Service{cfg: config, s: store, r: requests}
}

// Add creates song in library.
func (s *Service) Add(ctx context.Context, r models.RequestAddSong) (models.Song, error) {
	d, err := s.r.GetSongDetail(ctx, r)
	if err != nil {
		logger.Log.Info("unable to get song detail", zap.Error(err))
		return models.Song{}, status.Error(codes.Internal, "internal")
	}
	logger.Log.Debug("get detail success", zap.String("song", r.Song))

	ReleaseDateDate, ee := time.Parse("02.01.2006", d.ReleaseDate) // dd.mm.yyyy
	if ee != nil {
		logger.Log.Info("unable to parse release date", zap.String("releaseDate", d.ReleaseDate))
		return models.Song{}, status.Error(codes.Internal, "internal")
	}
	song, err := s.s.Add(ctx, models.Song{
		Group:       r.Group,
		Song:        r.Song,
		Text:        d.Text,
		ReleaseDate: ReleaseDateDate,
		Link:        d.Link})

	if err != nil {
		if err == storage.ErrUniqueViolation {
			return song, status.Error(codes.AlreadyExists, "already exists")
		}
		if err == sql.ErrNoRows {
			return models.Song{}, status.Error(codes.NotFound, "add deleted song")
		}
		logger.Log.Info("failed add song", zap.Error(err))
		return models.Song{}, status.Error(codes.Internal, "internal")
	}

	return song, nil
}

// Update changes song in library.
func (s *Service) Update(ctx context.Context, id string,
	data models.RequestUpdateSong) error {
	if err := s.s.Update(ctx, id, data); err != nil {
		if err == storage.ErrNotAffected {
			return status.Error(codes.NotFound, "not found")
		}
		return status.Error(codes.Internal, "internal")
	}
	return nil
}

// Delete removes song from library.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.s.Delete(ctx, id); err != nil {
		if err == storage.ErrNotAffected {
			return status.Error(codes.NotFound, "not found")
		}
		return status.Error(codes.Internal, "internal")
	}
	return nil
}

const (
	// DefaultPage is page by default.
	DefaultPage = 1
	// DefaultSizeText is size by default for song text pagination.
	DefaultSizeText = 3
	// DefaultSizeSongs is size by default for songs list pagination.
	DefaultSizeSongs = 10
)

// GetText returns song lyrics paginates by verses.
func (s *Service) GetText(ctx context.Context, id string,
	page, size int) (models.ResponseGetSongText, error) {
	d, err := s.s.GetText(ctx, id, page, size)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ResponseGetSongText{}, status.Error(codes.NotFound, "not found")
		}
		return d, status.Error(codes.Internal, "internal")
	}
	return d, nil
}

// GetSongs filters, paginates and returns library songs.
func (s *Service) GetSongs(ctx context.Context, d models.Song,
	page, size int) (models.ResponseGetSongs, error) {
	dd, err := s.s.GetSongs(ctx, d, page, size)
	if err != nil {
		return dd, status.Error(codes.Internal, "internal")
	}
	return dd, nil
}

// Ping checks storage availability.
func (s *Service) Ping() error { return s.s.Ping() }
