package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type ServerCfgOptions struct {
	port int
}
type server struct {
	cfg     ServerCfgOptions
	logger  zerolog.Logger
	handler http.Handler
}

// create a new instance of the server
func NewServer(logger zerolog.Logger, handlers []http.Handler, filePath string, port int) (server, error) {
	if filePath == "" {
		return server{}, errors.New("no file path provided")
	}

	// set the router DB
	router := newRouter(logger)
	err := router.setConn(filePath)
	if err != nil {
		return server{}, fmt.Errorf("error starting server database: %w", err)
	}

	muxHandler := http.NewServeMux()

	// set the router
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.set).Methods("POST")
	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.get).Methods("GET")
	muxRouter.HandleFunc("/{pk}/{project}", router.list).Methods("GET")
	muxRouter.HandleFunc("/{pk}/{project}", router.deleteProject).Methods("DELETE")
	muxRouter.HandleFunc("/{pk}/{project}/{key}", router.delete).Methods("DELETE")

	muxHandler.Handle("/", muxRouter)
	for _, handler := range handlers {
		muxHandler.Handle("/", handler)
	}

	cfg := ServerCfgOptions{
		port: port,
	}

	return server{
		logger:  logger,
		handler: muxHandler,
		cfg:     cfg,
	}, nil
}

func (s *server) Start() error {

	s.logger.Debug().Msg(fmt.Sprint("server is running at ", s.cfg.port))
	err := http.ListenAndServe(fmt.Sprintf(":%v", s.cfg.port), s.handler)

	if errors.Is(err, http.ErrServerClosed) {
		return errors.New("server closed")
	} else if err != nil {
		return fmt.Errorf("starting server failed with error: %w", err)
	}

	return nil
}
