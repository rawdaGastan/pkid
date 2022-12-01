package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// ServerCfgOptions is a struct for server configurations
type ServerCfgOptions struct {
	port int
}

// Server is a struct for server requirements
type Server struct {
	cfg     ServerCfgOptions
	logger  zerolog.Logger
	handler http.Handler
}

// NewServer creates a new instance of the server
func NewServer(logger zerolog.Logger, mws []mux.MiddlewareFunc, pkidStore PkidStore, filePath string, port int) (Server, error) {
	if filePath == "" {
		return Server{}, errors.New("no file path provided")
	}

	// set the router DB
	router := newRouter(logger, pkidStore)
	err := router.setConn(filePath)
	if err != nil {
		return Server{}, fmt.Errorf("error starting server database: %w", err)
	}

	muxHandler := http.NewServeMux()

	// set the router
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.set).Methods("POST")
	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.get).Methods("GET")
	muxRouter.HandleFunc("/{pk}/{project}", router.list).Methods("GET")
	muxRouter.HandleFunc("/{pk}/{project}", router.deleteProject).Methods("DELETE")
	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.delete).Methods("DELETE")

	for _, mw := range mws {
		muxRouter.Use(mw)
	}
	muxHandler.Handle("/", muxRouter)

	cfg := ServerCfgOptions{
		port: port,
	}

	return Server{
		logger:  logger,
		handler: muxHandler,
		cfg:     cfg,
	}, nil
}

// Start starts the server for the given server port
func (s *Server) Start() error {

	s.logger.Debug().Msg(fmt.Sprint("server is running at ", s.cfg.port))
	err := http.ListenAndServe(fmt.Sprintf(":%v", s.cfg.port), s.handler)

	if errors.Is(err, http.ErrServerClosed) {
		return errors.New("server closed")
	} else if err != nil {
		return fmt.Errorf("starting server failed with error: %w", err)
	}

	return nil
}
