// Package app for pkid app
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/pkid/config"
	"github.com/rawdaGastan/pkid/middlewares"
	"github.com/rawdaGastan/pkid/store"
	"github.com/rs/zerolog/log"
)

// App for all dependencies of backend server
type App struct {
	config config.Configuration
	db     store.PkidStore
}

// NewApp creates new server app all configurations
func NewApp(ctx context.Context, configFile string) (app *App, err error) {
	config, err := config.ReadConfFile(configFile)
	if err != nil {
		return
	}

	pkidStore := store.NewSqliteStore()
	err = pkidStore.SetConn(config.DBFile)
	if err != nil {
		return
	}

	if err = pkidStore.Migrate(); err != nil {
		return
	}

	return &App{
		config: config,
		db:     pkidStore,
	}, nil
}

// Start starts the app
func (a *App) Start(ctx context.Context) (err error) {
	a.registerHandlers()
	log.Info().Msgf("Server is listening on port %s", a.config.Port)

	srv := &http.Server{
		Addr: a.config.Port,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
		log.Info().Msg("Stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("HTTP shutdown error")
	}
	log.Info().Msg("Graceful shutdown complete")

	return nil
}

func (a *App) registerHandlers() {
	r := mux.NewRouter()

	versionRouter := r.PathPrefix("/" + a.config.Version).Subrouter()

	versionRouter.HandleFunc("/{pk}/{project}/{key}", WrapFunc(a.set)).Methods("POST", "OPTIONS")
	versionRouter.HandleFunc("/{pk}/{project}/{key}", WrapFunc(a.get)).Methods("GET", "OPTIONS")
	versionRouter.HandleFunc("/{pk}/{project}", WrapFunc(a.list)).Methods("GET", "OPTIONS")
	versionRouter.HandleFunc("/{pk}/{project}", WrapFunc(a.deleteProject)).Methods("DELETE", "OPTIONS")
	versionRouter.HandleFunc("/{pk}/{project}/{key}", WrapFunc(a.delete)).Methods("DELETE", "OPTIONS")

	// middlewares
	r.Use(middlewares.EnableCors)
	http.Handle("/", r)
}
