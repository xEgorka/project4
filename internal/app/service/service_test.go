package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/mocks"
	"github.com/xEgorka/project4/internal/app/models"
	"github.com/xEgorka/project4/internal/app/requests"
	"github.com/xEgorka/project4/internal/app/storage"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type args struct {
		config *config.Config
		store  storage.Storage
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{{name: "positive test #1"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			got := New(cfg, ms, requests.New(cfg))
			if reflect.TypeOf(got) == reflect.TypeOf((*Service)(nil)).Elem() {
				t.Errorf("not service")
			}
		})
	}
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	type args struct {
		ctx  context.Context
		song models.RequestAddSong
	}
	releaseDateStr := "16.07.2006"
	releaseDate, ee := time.Parse("02.01.2006", releaseDateStr)
	if ee != nil {
		panic(ee)
	}
	song := models.RequestAddSong{Group: "Muse", Song: "Supermassive Black Hole"}
	d := models.ResponseDetailSong{
		ReleaseDate: releaseDateStr,
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}
	ss := models.Song{
		Group:       song.Group,
		Song:        song.Song,
		ReleaseDate: releaseDate,
		Text:        d.Text,
		Link:        d.Link}
	r := args{ctx: context.Background(), song: song}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "positive test #1", args: r},
		{name: "negative test #1", args: r, wantErr: true},
		{name: "negative test #2", args: r, wantErr: true},
		{name: "negative test #3", args: r, wantErr: true},
		{name: "negative test #4", args: r, wantErr: true},
		{name: "negative test #5", args: r, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if err := json.NewEncoder(w).Encode(&d); err != nil {
							panic(err)
						}
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				ms.EXPECT().Add(tt.args.ctx, ss).Return(ss, nil)
				s.Add(tt.args.ctx, tt.args.song)
			}
			if tt.name == "negative test #1" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				if _, err := s.Add(tt.args.ctx, tt.args.song); (err != nil) != tt.wantErr {
					t.Errorf("HTTP.GetSongDetail() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if tt.name == "negative test #2" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if err := json.NewEncoder(w).Encode(&d); err != nil {
							panic(err)
						}
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				ms.EXPECT().Add(tt.args.ctx, ss).Return(ss, storage.ErrUniqueViolation)
				s.Add(tt.args.ctx, tt.args.song)
			}
			if tt.name == "negative test #3" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if err := json.NewEncoder(w).Encode(&d); err != nil {
							panic(err)
						}
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				ms.EXPECT().Add(tt.args.ctx, ss).Return(ss, sql.ErrNoRows)
				s.Add(tt.args.ctx, tt.args.song)
			}
			if tt.name == "negative test #4" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if err := json.NewEncoder(w).Encode(&d); err != nil {
							panic(err)
						}
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				ms.EXPECT().Add(tt.args.ctx, ss).Return(ss, errors.New("test"))
				s.Add(tt.args.ctx, tt.args.song)
			}
			if tt.name == "negative test #5" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						d.ReleaseDate = "bad"
						if err := json.NewEncoder(w).Encode(&d); err != nil {
							panic(err)
						}
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				s := New(cfg, ms, requests.New(cfg))
				s.Add(tt.args.ctx, tt.args.song)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	type args struct {
		ctx context.Context
		id  string
		d   models.RequestUpdateSong
	}
	releaseDateStr := "16.07.2006"
	releaseDate, ee := time.Parse("02.01.2006", releaseDateStr)
	if ee != nil {
		panic(ee)
	}
	id := "0824f9fb-7397-4f19-95d5-f9ce8bec75de"
	d := models.RequestUpdateSong{
		ReleaseDate: releaseDate,
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw"}
	cfg := &config.Config{}
	s := New(cfg, ms, requests.New(cfg))
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "positive test #1", args: args{ctx: context.Background(), id: id, d: d}},
		{name: "negative test #1", args: args{ctx: context.Background(), id: id, d: d}, wantErr: true},
		{name: "negative test #2", args: args{ctx: context.Background(), id: id, d: d}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
				ms.EXPECT().Update(tt.args.ctx, tt.args.id, tt.args.d).Return(nil)
				s.Update(tt.args.ctx, tt.args.id, tt.args.d)
			}
			if tt.name == "negative test #1" {
				ms.EXPECT().Update(tt.args.ctx, tt.args.id, tt.args.d).Return(storage.ErrNotAffected)
				s.Update(tt.args.ctx, tt.args.id, tt.args.d)
			}
			if tt.name == "negative test #2" {
				ms.EXPECT().Update(tt.args.ctx, tt.args.id, tt.args.d).Return(errors.New("test"))
				s.Update(tt.args.ctx, tt.args.id, tt.args.d)
			}

		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	type args struct {
		ctx context.Context
		id  string
	}
	id := "0824f9fb-7397-4f19-95d5-f9ce8bec75de"
	cfg := &config.Config{}
	s := New(cfg, ms, requests.New(cfg))
	r := args{ctx: context.Background(), id: id}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "positive test #1", args: r},
		{name: "negative test #1", args: r, wantErr: true},
		{name: "negative test #2", args: r, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
				ms.EXPECT().Delete(tt.args.ctx, tt.args.id).Return(nil)
				s.Delete(tt.args.ctx, tt.args.id)
			}
			if tt.name == "negative test #1" {
				ms.EXPECT().Delete(tt.args.ctx, tt.args.id).Return(storage.ErrNotAffected)
				s.Delete(tt.args.ctx, tt.args.id)
			}
			if tt.name == "negative test #2" {
				ms.EXPECT().Delete(tt.args.ctx, tt.args.id).Return(errors.New("test"))
				s.Delete(tt.args.ctx, tt.args.id)
			}

		})
	}
}

func TestGetText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	type args struct {
		ctx context.Context
		id  string
	}
	id := "0824f9fb-7397-4f19-95d5-f9ce8bec75de"
	page := 1
	size := 3
	cfg := &config.Config{}
	s := New(cfg, ms, requests.New(cfg))
	r := args{ctx: context.Background(), id: id}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "positive test #1", args: r},
		{name: "negative test #1", args: r, wantErr: true},
		{name: "negative test #2", args: r, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
				ms.EXPECT().GetText(tt.args.ctx, tt.args.id, page, size).
					Return(models.ResponseGetSongText{}, nil)
				s.GetText(tt.args.ctx, tt.args.id, page, size)
			}
			if tt.name == "negative test #1" {
				ms.EXPECT().GetText(tt.args.ctx, tt.args.id, page, size).
					Return(models.ResponseGetSongText{}, sql.ErrNoRows)
				s.GetText(tt.args.ctx, tt.args.id, page, size)
			}
			if tt.name == "negative test #2" {
				ms.EXPECT().GetText(tt.args.ctx, tt.args.id, page, size).
					Return(models.ResponseGetSongText{}, errors.New("test"))
				s.GetText(tt.args.ctx, tt.args.id, page, size)
			}

		})
	}
}

func TestSongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	type args struct {
		ctx context.Context
		req models.Song
	}
	req := models.Song{
		ID: "0824f9fb-7397-4f19-95d5-f9ce8bec75de",
	}
	cfg := &config.Config{}
	s := New(cfg, ms, requests.New(cfg))
	r := args{ctx: context.Background(), req: req}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "positive test #1", args: r},
		{name: "negative test #1", args: r, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
				ms.EXPECT().GetSongs(tt.args.ctx, tt.args.req, DefaultPage, DefaultSizeSongs).
					Return(models.ResponseGetSongs{}, nil)
				s.GetSongs(tt.args.ctx, tt.args.req, DefaultPage, DefaultSizeSongs)
			}
			if tt.name == "negative test #1" {
				ms.EXPECT().GetSongs(tt.args.ctx, tt.args.req, DefaultPage, DefaultSizeSongs).
					Return(models.ResponseGetSongs{}, errors.New("test"))
				s.GetSongs(tt.args.ctx, tt.args.req, DefaultPage, DefaultSizeSongs)
			}

		})
	}
}

func TestService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := New(cfg, ms, requests.New(cfg))
	type fields struct {
		cfg *config.Config
		s   storage.Storage
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "positive test #1",
			wantErr: false,
		},
		{
			name:    "negative test #1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if tt.name == "positive test #1" {
			ms.EXPECT().Ping().Return(nil)
		}
		if tt.name == "negative test #1" {
			ms.EXPECT().Ping().Return(errors.New("test"))
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("Service.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
