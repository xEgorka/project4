package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/mocks"
	"github.com/xEgorka/project4/internal/app/models"
	"github.com/xEgorka/project4/internal/app/requests"
	"github.com/xEgorka/project4/internal/app/service"
	"github.com/xEgorka/project4/internal/app/storage"
)

func TestHTTP_PostSong(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			d := models.ResponseDetailSong{
				ReleaseDate: "16.07.2006",
				Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
				Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}
			if err := json.NewEncoder(w).Encode(&d); err != nil {
				panic(err)
			}
		}))
	defer func() { srv.Close() }()
	cfg := &config.Config{MusicInfoURL: srv.URL}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: `{"group": "Muse","song": "Supermassive Black Hole"}`,
			want: want{code: http.StatusOK, contentType: "application/json"},
		},
		{
			name: "negative test #1",
			body: `{"group": "Muse","song": ""}`,
			want: want{code: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "negative test #2",
			body: `bad json`,
			want: want{code: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "negative test #3",
			body: `{"group": "Muse","song": "Supermassive Black Hole"}`,
			want: want{code: http.StatusInternalServerError},
		},
		{
			name: "negative test #4",
			body: `{"group": "Muse","song": "Supermassive Black Hole"}`,
			want: want{code: http.StatusConflict},
		},
		{
			name: "negative test #5",
			body: `{"group": "Muse","song": "Supermassive Black Hole"}`,
			want: want{code: http.StatusGone},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/song", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctx := context.Background()
			releaseDate, ee := time.Parse("02.01.2006", "16.07.2006")
			if ee != nil {
				panic(ee)
			}
			s := models.Song{
				Group:       "Muse",
				Song:        "Supermassive Black Hole",
				ReleaseDate: releaseDate,
				Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
				Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}
			if tt.want.code == http.StatusOK {
				ms.EXPECT().Add(ctx, s).Return(s, nil)
			}
			if tt.name == "negative test #3" {
				ms.EXPECT().Add(ctx, s).Return(s, errors.New("test error"))
			}
			if tt.name == "negative test #4" {
				ms.EXPECT().Add(ctx, s).Return(s, storage.ErrUniqueViolation)
			}
			if tt.name == "negative test #5" {
				ms.EXPECT().Add(ctx, s).Return(s, sql.ErrNoRows)
			}
			h.PostSong(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestHTTP_UpdateSong(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		id   string
		body string
		want want
	}{
		{
			name: "positive test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			body: `{"release_date": "2006-07-16T00:00:00Z","text": "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight","link": "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}`,
			want: want{code: http.StatusAccepted},
		},
		{
			name: "negative test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			body: `{"bad"}`,
			want: want{contentType: "text/plain; charset=utf-8", code: http.StatusBadRequest},
		},
		{
			name: "negative test #2",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			body: `{"release_date": "2006-07-16T00:00:00Z"}`,
			want: want{contentType: "text/plain; charset=utf-8", code: http.StatusBadRequest},
		},
		{
			name: "negative test #3",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			body: `{"release_date": "2006-07-16T00:00:00Z","text": "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight","link": "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}`,
			want: want{code: http.StatusNoContent},
		},
		{
			name: "negative test #4",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			body: `{"release_date": "2006-07-16T00:00:00Z","text": "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight","link": "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}`,
			want: want{code: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPut, "/api/song/{id}", strings.NewReader(tt.body))
			r.SetPathValue("id", tt.id)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctx := context.Background()
			releaseDate, ee := time.Parse("02.01.2006", "16.07.2006")
			if ee != nil {
				panic(ee)
			}
			s := models.RequestUpdateSong{
				ReleaseDate: releaseDate,
				Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
				Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}
			if tt.want.code == http.StatusAccepted {
				ms.EXPECT().Update(ctx, tt.id, s).Return(nil)
			}
			if tt.name == "negative test #3" {
				ms.EXPECT().Update(ctx, tt.id, s).Return(storage.ErrNotAffected)
			}
			if tt.name == "negative test #4" {
				ms.EXPECT().Update(ctx, tt.id, s).Return(errors.New("test"))
			}
			h.PutSong(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestHTTP_DeleteSong(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		id   string
		body string
		want want
	}{
		{
			name: "positive test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusAccepted},
		},
		{
			name: "negative test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusNoContent},
		},
		{
			name: "negative test #2",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodDelete, "/api/song/{id}", strings.NewReader(""))
			r.SetPathValue("id", tt.id)
			r.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			ctx := context.Background()

			if tt.want.code == http.StatusAccepted {
				ms.EXPECT().Delete(ctx, tt.id).Return(nil)
			}
			if tt.name == "negative test #1" {
				ms.EXPECT().Delete(ctx, tt.id).Return(storage.ErrNotAffected)
			}
			if tt.name == "negative test #2" {
				ms.EXPECT().Delete(ctx, tt.id).Return(errors.New("test"))
			}
			h.DeleteSong(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestHTTP_GetSongText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		id   string
		body string
		want want
	}{
		{
			name: "positive test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusOK, contentType: "application/json"},
		},
		{
			name: "negative test #1",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{contentType: "text/plain; charset=utf-8", code: http.StatusBadRequest},
		},
		{
			name: "negative test #2",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{contentType: "text/plain; charset=utf-8", code: http.StatusBadRequest},
		},
		{
			name: "negative test #3",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusNoContent},
		},
		{
			name: "negative test #4",
			id:   "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
			want: want{code: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/song/{id}/text", strings.NewReader(""))
			if tt.name == "negative test #1" {
				r = httptest.NewRequest(http.MethodGet, "/api/song/{id}/text?page=bad", strings.NewReader(""))
			}
			if tt.name == "negative test #2" {
				r = httptest.NewRequest(http.MethodGet, "/api/song/{id}/text?size=bad", strings.NewReader(""))
			}
			r.SetPathValue("id", tt.id)
			r.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			ctx := context.Background()

			s := models.ResponseGetSongText{
				ID:     "ca1da5fa-50ee-4d00-82e9-d6a578419ad7",
				Group:  "Muse",
				Page:   service.DefaultPage,
				Size:   service.DefaultSizeText,
				Song:   "Supermassive Black Hole",
				Total:  2,
				Verses: []string{"Ooh baby don't you know I suffer?\nOoh baby can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?", "Ooh\nYou set my soul alight\nOoh\nYou set my soul alight"},
			}
			if tt.want.code == http.StatusOK {
				ms.EXPECT().GetText(ctx, tt.id, service.DefaultPage, service.DefaultSizeText).Return(s, nil)
			}
			if tt.name == "negative test #3" {
				ms.EXPECT().GetText(ctx, tt.id, service.DefaultPage, service.DefaultSizeText).Return(models.ResponseGetSongText{}, sql.ErrNoRows)
			}
			if tt.name == "negative test #4" {
				ms.EXPECT().GetText(ctx, tt.id, service.DefaultPage, service.DefaultSizeText).Return(models.ResponseGetSongText{}, errors.New("test"))
			}
			h.GetSongText(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestHTTP_GetSongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name string
		id   string
		body string
		want want
	}{
		{name: "positive test #1", want: want{code: http.StatusOK, contentType: "application/json"}},
		{name: "negative test #1", want: want{code: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "negative test #2", want: want{code: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "negative test #3", want: want{code: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "negative test #4", want: want{code: http.StatusInternalServerError, contentType: "text/plain; charset=utf-8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/songs", strings.NewReader(""))
			if tt.name == "negative test #1" {
				r = httptest.NewRequest(http.MethodGet, "/api/songs?size=bad", strings.NewReader(""))
			}
			if tt.name == "negative test #2" {
				r = httptest.NewRequest(http.MethodGet, "/api/songs?page=0", strings.NewReader(""))
			}
			if tt.name == "negative test #3" {
				r = httptest.NewRequest(http.MethodGet, "/api/songs?release_date=bad", strings.NewReader(""))
			}
			r.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			ctx := context.Background()
			var releaseDate time.Time
			req := models.Song{ReleaseDate: releaseDate}
			s := models.ResponseGetSongs{
				Page:  service.DefaultPage,
				Size:  service.DefaultSizeSongs,
				Songs: []models.Song{},
			}
			if tt.want.code == http.StatusOK {
				ms.EXPECT().GetSongs(ctx, req, service.DefaultPage, service.DefaultSizeSongs).Return(s, nil)
			}
			if tt.name == "negative test #4" {
				ms.EXPECT().GetSongs(ctx, req, service.DefaultPage, service.DefaultSizeSongs).Return(s, errors.New("test"))
			}
			h.GetSongs(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestHandlers_GetPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(cfg, ms, requests.New(cfg))
	h := NewHTTP(s)
	type want struct {
		contentType string
		url         string
		code        int
	}
	tests := []struct {
		name      string
		body      string
		userID    string
		want      want
		timestamp int
	}{
		{
			name: "positive test #1",
			want: want{code: http.StatusOK, contentType: "text/plain"},
		},
		{
			name: "negative test #1",
			want: want{code: http.StatusInternalServerError, contentType: "text/plain; charset=utf-8"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ping",
				strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			ctx := context.Background()
			if tt.want.code == http.StatusOK {
				ms.EXPECT().Ping().Return(nil)
			} else {
				ms.EXPECT().Ping().Return(errors.New("test"))
			}
			h.GetPing(w, r.WithContext(ctx))
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
		})
	}
}
