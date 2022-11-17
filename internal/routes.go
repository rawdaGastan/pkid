package internal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

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
		s.logger.Error().Msg(fmt.Sprint("can't find key: ", err))
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"data": value, "msg": "data is got successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}

// list all keys for a specific project, using the public key
func (s *server) list(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	if project == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("db list project failed with error: no project given")
		return
	}

	AllKeys, err := s.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("db list failed with error: ", err))
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
		s.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}

func (s *server) deleteProject(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	if project == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg("db deleting project failed with error: no project given")
		return
	}

	AllKeys, err := s.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("db list failed with error: ", err))
		return
	}

	for _, key := range AllKeys {
		if strings.HasPrefix(key, pk+"_"+project+"_") {
			err := s.db.delete(key)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				s.logger.Error().Msg(fmt.Sprintf("db deleting key %v failed with error: %v", key, err))
				return
			}
		}
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
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
		s.logger.Error().Msg(fmt.Sprint("db deletion failed with error: ", err))
		return
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
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
		s.logger.Error().Msg(fmt.Sprint("can't decode public key: ", err))
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
	verified, err := verifySignedData(body, verifyPk)
	if !verified || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("invalid data: ", err))
		return
	}
	s.logger.Debug().Msg(fmt.Sprint("signed body is verified: ", verified))

	authHeader, err := verifySignedHeader(request.Header.Get("Authorization"), verifyPk)
	if !authHeader || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("invalid authorization header: ", err))
		return
	}
	s.logger.Debug().Msg(fmt.Sprint("signed header is verified: ", err))

	// set date
	docKey := pk + "_" + projectKey
	err = s.db.set(docKey, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("database set failed with error: ", err))
		return
	}

	// response
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is set successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}
