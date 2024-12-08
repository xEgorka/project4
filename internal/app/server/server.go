// Package server starts storage and defines HTTP router.
package server

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	_ "github.com/xEgorka/project4/swagger" // generated docs

	"github.com/xEgorka/project4/internal/app/config"
	"github.com/xEgorka/project4/internal/app/handlers"
	"github.com/xEgorka/project4/internal/app/logger"
	"github.com/xEgorka/project4/internal/app/requests"
	"github.com/xEgorka/project4/internal/app/service"
	"github.com/xEgorka/project4/internal/app/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// Start starts server.
func Start() error {
	w := bufio.NewWriter(os.Stdout)
	if _, err := fmt.Fprintf(w, "ONLINE SONG LIBRARY SERVER\nVersion: %s\t%s\nCommit: %s\n\n",
		buildVersion, buildDate, buildCommit); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if e := logger.Initialize("debug"); e != nil {
		return e
	}
	if err := godotenv.Load(); err != nil {
		return err
	}
	cfg, err := config.Setup()
	if err != nil {
		return err
	}
	ctx := context.Background()
	s, err := storage.Open(ctx, cfg)
	if err != nil {
		return err
	}
	srv := http.Server{
		Addr:    cfg.URI,
		Handler: routes(handlers.NewHTTP(service.New(cfg, s, requests.New(cfg)))),
	}

	go func() {
		logger.Log.Info("running http server...", zap.String("uri", cfg.URI))
		logger.Log.Info("music info api", zap.String("url", cfg.MusicInfoURL))
		logger.Log.Info("swagger address", zap.String("url", cfg.URI+"/swagger/index.html#/"))
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Log.Info("http server stopping... done")
			} else {
				logger.Log.Error("failed run http server", zap.Error(err))
			}
		}
	}()
	return stop(&srv)
}

var sigint = make(chan os.Signal, 1)

const timeout = 5 * time.Second

func stop(srv *http.Server) error {
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-sigint
	logger.Log.Info("signal received", zap.String("sig", sig.String()))

	logger.Log.Info("server stopping...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("failed server stop", zap.Error(err))
		return err
	}
	return nil
}

// @Title Online Song Library API
// @Description Online Song Library.
// @Version 0.1

// @Contact.email x0o1@ya.ru

// @BasePath /api
// @Host localhost:8080

// @Tag.name Songs
// @Tag.description "Songs requests group."
func routes(h handlers.HTTP) *chi.Mux {
	r := chi.NewRouter()
	r.Use(handlers.WithLogging)

	r.Get("/api/ping", h.GetPing)
	r.Post("/api/song", h.PostSong)
	r.Put("/api/song/{id}", h.PutSong)
	r.Delete("/api/song/{id}", h.DeleteSong)
	r.Get("/api/song/{id}/text", h.GetSongText)
	r.Get("/api/songs", h.GetSongs)

	r.Get("/swagger/*",
		httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	return r
}
