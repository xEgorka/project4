package server

import (
	"context"
	"net/http"
	"reflect"
	"syscall"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/handlers"
	"github.com/xEgorka/project4/internal/app/mocks"
	"github.com/xEgorka/project4/internal/app/requests"
	"github.com/xEgorka/project4/internal/app/service"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{{name: "negative test #1", wantErr: true}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stop(t *testing.T) {
	_, cancelBatch := context.WithCancel(context.Background())
	srv := http.Server{}
	go srv.ListenAndServe()
	type args struct {
		cancelBatch context.CancelFunc
		srv         *http.Server
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "positive test #1",
			args:    args{cancelBatch: cancelBatch, srv: &srv},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				if err := stop(tt.args.srv); (err != nil) != tt.wantErr {
					t.Errorf("stop() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			sigint <- syscall.SIGQUIT
		})
	}
}

func Test_routes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ms := mocks.NewMockStorage(ctrl)
	cfg := &config.Config{}
	s := service.New(&config.Config{}, ms, requests.New(cfg))
	h := handlers.NewHTTP(s)
	type args struct {
		h handlers.HTTP
	}
	tests := []struct {
		name string
		args args
		want *chi.Mux
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routes(h)
			if reflect.TypeOf(got) == reflect.TypeOf((*chi.Mux)(nil)).Elem() {
				t.Errorf("not chi mux")
			}
		})
	}
}
