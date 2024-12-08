package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xEgorka/project4/internal/app/logger"
	"github.com/xEgorka/project4/internal/app/models"
	"github.com/xEgorka/project4/internal/app/service"
)

// HTTP provides methods for http server.
type HTTP struct{ s *service.Service }

// NewHTTP creates HTTP.
func NewHTTP(service *service.Service) HTTP { return HTTP{s: service} }

// PostSong godoc
// @Summary Add song
// @Description Add song to library
// @Tags Songs
// @Accept json
// @Produce json
// @Param song body models.RequestAddSong true "Add song"
// @Success 200 {object} models.Song "Song added"
// @Failure 400 "Bad request"
// @Failure 409 "Song already exists"
// @Failure 410 "Song already deleted"
// @Failure 500 "Internal server error"
// @Router /song [post]
func (h *HTTP) PostSong(w http.ResponseWriter, r *http.Request) {
	var req models.RequestAddSong
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Info("JSON decode error", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if len(req.Group) == 0 || len(req.Song) == 0 {
		logger.Log.Info("empty group or song")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	d, err := h.s.Add(r.Context(), req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.AlreadyExists:
				w.WriteHeader(http.StatusConflict)
			case codes.NotFound:
				w.WriteHeader(http.StatusGone)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(&d); err != nil {
		logger.Log.Info("JSON encode error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// PutSong godoc
// @Summary Update song
// @Description Update song in library
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song id"
// @Param song body models.RequestUpdateSong true "Update song"
// @Success 202 "Song updated"
// @Failure 204 "Song not found"
// @Failure 400 "Bad request"
// @Failure 500 "Internal server error"
// @Router /song/{id} [put]
func (h *HTTP) PutSong(w http.ResponseWriter, r *http.Request) {
	var req models.RequestUpdateSong
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Info("JSON decode error", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.ReleaseDate.IsZero() || len(req.Text) == 0 || len(req.Link) == 0 {
		logger.Log.Info("empty release date, text or link")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.s.Update(r.Context(), r.PathValue("id"), req); err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				w.WriteHeader(http.StatusNoContent)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

// DeleteSong godoc
// @Summary Delete song
// @Description Delete song from library
// @Tags Songs
// @Param id path string true "Song id"
// @Success 202 "Song deleted"
// @Failure 204 "Song not found"
// @Failure 500 "Internal server error"
// @Router /song/{id} [delete]
func (h *HTTP) DeleteSong(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Delete(r.Context(), r.PathValue("id")); err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				w.WriteHeader(http.StatusNoContent)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

// GetSongText godoc
// @Summary Get song text
// @Description Get song text for certain page and page size
// @Tags Songs
// @Produce json
// @Param id path string true "Song id"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(3)
// @Success 200 {object} models.ResponseGetSongText "Song text"
// @Failure 204 "Song not found"
// @Failure 400 "Bad request"
// @Failure 500 "Internal server error"
// @Router /song/{id}/text [get]
func (h *HTTP) GetSongText(w http.ResponseWriter, r *http.Request) {
	var page, size int
	var err error
	pageStr, sizeStr := r.URL.Query().Get("page"), r.URL.Query().Get("size")
	if len(pageStr) == 0 {
		page = service.DefaultPage
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			logger.Log.Info("invalid page")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	if len(sizeStr) == 0 {
		size = service.DefaultSizeText
	} else {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			logger.Log.Info("invalid size")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	d, err := h.s.GetText(r.Context(), r.PathValue("id"), page, size)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				w.WriteHeader(http.StatusNoContent)
			case codes.Internal:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(&d); err != nil {
		logger.Log.Info("JSON encode error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetSongs godoc
// @Summary Get songs
// @Description Get filtered songs list for certain page and page size
// @Tags Songs
// @Produce json
// @Param id query string false "Song id"
// @Param group query string false "Group"
// @Param song query string false "Song"
// @Param release_date query string false "Release date" default(16.07.2006)
// @Param text query string false "Text"
// @Param link query string false "Link"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} models.ResponseGetSongs "Songs list"
// @Failure 400 "Bad request"
// @Failure 500 "Internal server error"
// @Router /songs [get]
func (h *HTTP) GetSongs(w http.ResponseWriter, r *http.Request) {
	var page, size int
	var err error
	pageStr, sizeStr := r.URL.Query().Get("page"), r.URL.Query().Get("size")
	if len(pageStr) == 0 {
		page = service.DefaultPage
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			logger.Log.Info("invalid page")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	if len(sizeStr) == 0 {
		size = service.DefaultSizeSongs
	} else {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			logger.Log.Info("invalid size")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	releaseDateStr := r.URL.Query().Get("release_date")
	var releaseDate time.Time
	var ee error
	if len(releaseDateStr) > 0 {
		releaseDate, ee = time.Parse("02.01.2006", releaseDateStr)
		if ee != nil {
			logger.Log.Info("invalid release date")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	d, err := h.s.GetSongs(r.Context(), models.Song{
		ID:          r.URL.Query().Get("id"),
		Group:       r.URL.Query().Get("group"),
		Song:        r.URL.Query().Get("song"),
		ReleaseDate: releaseDate,
		Text:        r.URL.Query().Get("text"),
		Link:        r.URL.Query().Get("link"),
	}, page, size)
	if err != nil {
		logger.Log.Info("unable to get songs", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(&d); err != nil {
		logger.Log.Info("JSON encode error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetPing checks service availability.
func (h *HTTP) GetPing(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Ping(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "text/plain")
}
