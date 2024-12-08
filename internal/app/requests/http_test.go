package requests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/models"
)

func TestHTTP_GetSongDetail(t *testing.T) {
	type fields struct {
		cfg *config.Config
		c   *http.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		srv     *httptest.Server
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "positive test #1",
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "negative test #1",
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "negative test #2",
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "positive test #1" {
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
				r := New(cfg)
				if _, err := r.GetSongDetail(context.Background(), models.RequestAddSong{Group: "Muse", Song: "Supermassive Black Hole"}); (err != nil) != tt.wantErr {
					t.Errorf("HTTP.Register() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if tt.name == "negative test #1" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-type", "text/plain")
						w.WriteHeader(http.StatusBadRequest)
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				r := New(cfg)
				if _, err := r.GetSongDetail(context.Background(), models.RequestAddSong{Group: "Muse", Song: "Supermassive Black Hole"}); (err != nil) != tt.wantErr {
					t.Errorf("HTTP.Register() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if tt.name == "negative test #2" {
				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)
					}))
				defer func() { srv.Close() }()
				cfg := &config.Config{MusicInfoURL: srv.URL}
				r := New(cfg)
				if _, err := r.GetSongDetail(context.Background(), models.RequestAddSong{Group: "Muse", Song: "Supermassive Black Hole"}); (err != nil) != tt.wantErr {
					t.Errorf("HTTP.Register() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
