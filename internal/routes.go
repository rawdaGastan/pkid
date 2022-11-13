package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type server struct {
	db *SQLiteDB
}

// create a new instance of the server
func newServer() server {
	return server{}
}

// set the connection and migration of the db
func (s *server) setConn(filename string) error {
	db := newSQLiteDB()
	db.setConn(filename)
	if err := db.migrate(); err != nil {
		return err
	}

	s.db = db
	return nil
}

// get the value of the given key, using the public key
func (s *server) get(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	key = project + "_" + key

	docKey := hex.EncodeToString(sha256.New().Sum([]byte(pk + key)))

	value, err := s.db.get(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "can't find key: %v\n", err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "response failed with error: %v\n", err)
		return
	}
	w.Write(res)
}

/*
// list all keys for a specific project, using the public key
func (s *server) list(w http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(w, "start\n")

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	// verify key
	verifyPk, err := hex.DecodeString(pk)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "can't decode public key: %v\n", err)
		return
	}

	allKeys, err := s.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "list failed: %v\n", err)
		return
	}

	keys := []string{}
	for _, k := range allKeys {
		//docKey := hex.EncodeToString(sha256.New().Sum([]byte(pk + key)))
		decoded, err := hex.DecodeString(k)
		if err != nil {
			fmt.Fprintf(w, "%v\n", err)
			continue
		}

		sha256.New().Sum([]byte(pk + key))

		unsignedKey, err := unsign(k, verifyPk)
		if err != nil {
			fmt.Fprintf(w, "%v\n", err)
			continue
		}

		fmt.Fprintf(w, "%v\n", unsignedKey)
		if strings.Contains(string(unsignedKey), pk+project) {
			keys = append(keys, string(unsignedKey))
		}
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%v\n", keys)
}
*/

// delete the value of the given key, using the public key
func (s *server) delete(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	key = project + "_" + key

	docKey := hex.EncodeToString(sha256.New().Sum([]byte(pk + key)))

	err := s.db.delete(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "deletion error: %v\n", err)
		return
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"message": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "response failed with error: %v\n", err)
		return
	}
	w.Write(res)

}

// set the given value of the given key, using the public key
func (s *server) set(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	key = project + "_" + key

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	body := buf.String()

	// verify key
	verifyPk, err := hex.DecodeString(pk)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "can't decode public key: %v\n", err)
		return
	}

	// check request
	if request.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "no body given\n")
		return
	}

	if request.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "no Authorization header\n")
		return
	}

	// verify
	data, err := verifySignedData(w, string(body), verifyPk)
	if !data || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid data: %v\n", err)
		return
	}

	authHeader, err := verifySignedHeader(w, request.Header.Get("Authorization"), verifyPk)
	if !authHeader || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid authorization header %v\n", err)
		return
	}

	// set date
	docKey := hex.EncodeToString(sha256.New().Sum([]byte(pk + key)))
	err = s.db.set(docKey, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "database ser failed with error: %v\n", err)
		return
	}

	// response
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"message": "data is set successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "response failed with error: %v\n", err)
		return
	}
	w.Write(res)
}
