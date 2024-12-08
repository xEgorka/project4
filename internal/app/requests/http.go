// Package requests makes http requests.
package requests

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/logger"
	"github.com/xEgorka/project4/internal/app/models"
)

// HTTP provides methods for http requests.
type HTTP struct {
	cfg *config.Config
	c   *http.Client
}

func newClient() *http.Client { return &http.Client{} }

// New creates HTTP.
func New(config *config.Config) *HTTP { return &HTTP{cfg: config, c: newClient()} }

// GetSongDetail requests song details.
func (h *HTTP) GetSongDetail(ctx context.Context, d models.RequestAddSong) (
	models.ResponseDetailSong, error) {
	var s models.ResponseDetailSong

	url := h.cfg.MusicInfoURL + "/info"
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return s, err
	}

	q := r.URL.Query()
	q.Add("group", d.Group)
	q.Add("song", d.Song)
	r.URL.RawQuery = q.Encode()

	res, ee := h.c.Do(r)
	if ee != nil {
		return s, ee
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Log.Info("failed body close", zap.Error(err))
		}
	}()

	var e error
	switch {
	case res.StatusCode == http.StatusOK:
		if e = json.NewDecoder(res.Body).Decode(&s); e != nil {
			return s, e
		}
		return s, nil
	case res.StatusCode == http.StatusBadRequest:
		e = status.Error(codes.InvalidArgument, "invalid argument")
	case res.StatusCode == http.StatusInternalServerError:
		e = status.Error(codes.Internal, "internal")
	}
	return s, e
}
