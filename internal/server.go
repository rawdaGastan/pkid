package internal

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func StartServer(filename string, port string) {
	// set the server
	server := newServer()
	err := server.setConn(filename)
	if err != nil {
		fmt.Print("error starting server database: \n", err)
		os.Exit(1)
	}

	// set the router
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/set/{pk}/{project}/{key}", server.set).Methods("POST")
	router.HandleFunc("/get/{pk}/{project}/{key}", server.get).Methods("GET")
	//router.HandleFunc("/list/{pk}/{project}", server.list).Methods("GET")
	router.HandleFunc("/delete/{pk}/{project}/{key}", server.delete).Methods("DELETE")

	// run server
	fmt.Println("server is running at", port)
	err = http.ListenAndServe(":"+port, router)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Print("server closed\n")
	} else if err != nil {
		fmt.Print("error starting server: \n", err)
		os.Exit(1)
	}
}
