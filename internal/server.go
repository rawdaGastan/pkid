package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

func StartServer(logger zerolog.Logger, filePath string, port int) error {

	if filePath == "" {
		return errors.New("no file path provided")
	}

	// set the server
	server := newServer(logger)
	err := server.setConn(filePath)
	if err != nil {
		return errors.New("error starting server database: " + fmt.Sprint(err))
	}

	// set the router
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/set/{pk}/{project}/{key}", server.set).Methods("POST")
	router.HandleFunc("/get/{pk}/{project}/{key}", server.get).Methods("GET")
	router.HandleFunc("/list/{pk}/{project}", server.list).Methods("GET")
	router.HandleFunc("/delete/{pk}/{project}/{key}", server.delete).Methods("DELETE")

	// start the server
	logger.Debug().Msg("server is running at " + fmt.Sprint(port))
	err = http.ListenAndServe(":"+fmt.Sprint(port), router)

	if errors.Is(err, http.ErrServerClosed) {
		return errors.New("server closed")
	} else if err != nil {
		return errors.New("starting server failed with error: " + fmt.Sprint(err))
	}

	return nil
}
