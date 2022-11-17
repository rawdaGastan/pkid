package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type server struct {
	db     PkidStore
	logger zerolog.Logger
}

// create a new instance of the server
func newServer(logger zerolog.Logger) server {
	return server{
		logger: logger,
	}
}

// set the connection and migration of the db
func (s *server) setConn(filePath string) error {
	if filePath == "" {
		return errors.New("no file path provided")
	}

	db := newPkidStore()
	db.setConn(filePath)
	if err := db.migrate(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func StartServer(logger zerolog.Logger, filePath string, port int) error {

	if filePath == "" {
		return errors.New("no file path provided")
	}

	// set the server
	server := newServer(logger)
	err := server.setConn(filePath)
	if err != nil {
		return fmt.Errorf("error starting server database: %w", err)
	}

	// set the router
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/{pk}/{project}/{key}", server.set).Methods("POST")
	router.HandleFunc("/{pk}/{project}/{key}", server.get).Methods("GET")
	router.HandleFunc("/{pk}/{project}", server.list).Methods("GET")
	router.HandleFunc("/{pk}/{project}", server.deleteProject).Methods("DELETE")
	router.HandleFunc("/{pk}/{project}/{key}", server.delete).Methods("DELETE")

	// start the server
	logger.Debug().Msg(fmt.Sprint("server is running at ", port))
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), router)

	if errors.Is(err, http.ErrServerClosed) {
		return errors.New("server closed")
	} else if err != nil {
		return fmt.Errorf("starting server failed with error: %w", err)
	}

	return nil
}
