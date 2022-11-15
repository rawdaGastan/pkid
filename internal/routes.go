package internal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type server struct {
	db     *sqliteDB
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

	db := newSQLiteDB()
	db.setConn(filePath)
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
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	value, err := s.db.get(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("can't find key: " + fmt.Sprint(err))
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"data": value, "msg": "data is got successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("response failed with error: " + fmt.Sprint(err))
		return
	}
	w.Write(res)
}

// list all keys for a specific project, using the public key
func (s *server) list(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	AllKeys, err := s.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("db list failed with error: " + fmt.Sprint(err))
		return
	}

	keys := []string{}
	for _, key := range AllKeys {
		if strings.HasPrefix(key, pk+"_"+project+"_") {
			splitKey := strings.Split(key, "_")
			if len(splitKey) == 3 {
				keys = append(keys, splitKey[2])
			}
		}
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]any{"data": keys, "msg": "data is got successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("response failed with error: " + fmt.Sprint(err))
		return
	}
	w.Write(res)
}

// delete the value of the given key, using the public key
func (s *server) delete(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	err := s.db.delete(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("db deletion failed with error: " + fmt.Sprint(err))
		return
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("response failed with error: " + fmt.Sprint(err))
		return
	}
	w.Write(res)

}

// set the given value of the given key, using the public key
func (s *server) set(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	projectKey := project + "_" + key

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	body := buf.String()

	// verify key
	verifyPk, err := hex.DecodeString(pk)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("can't decode public key: " + fmt.Sprint(err))
		return
	}

	// check request
	if body == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("no body given")
		return
	}

	if request.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("no Authorization header")
		return
	}

	// verify
	verified, err := verifySignedData(s.logger, body, verifyPk)
	if !verified || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("invalid data: " + fmt.Sprint(err))
		return
	}

	authHeader, err := verifySignedHeader(s.logger, request.Header.Get("Authorization"), verifyPk)
	if !authHeader || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("invalid authorization header: " + fmt.Sprint(err))
		return
	}

	// set date
	docKey := pk + "_" + projectKey
	err = s.db.set(docKey, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("database set failed with error: " + fmt.Sprint(err))
		return
	}

	// response
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is set successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("response failed with error: " + fmt.Sprint(err))
		return
	}
	w.Write(res)
}
