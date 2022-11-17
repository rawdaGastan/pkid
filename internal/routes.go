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

type router struct {
	db     PkidStore
	logger zerolog.Logger
}

// create a new instance of the router
func newRouter(logger zerolog.Logger) router {
	return router{
		logger: logger,
	}
}

// set the connection and migration of the db
func (r *router) setConn(filePath string) error {
	if filePath == "" {
		return errors.New("no file path provided")
	}

	db := newPkidStore()
	db.setConn(filePath)
	if err := db.migrate(); err != nil {
		return err
	}

	r.db = db
	return nil
}

// get the value of the given key, using the public key
func (r *router) get(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	value, err := r.db.get(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("can't find key: ", err))
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"data": value, "msg": "data is got successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}

// list all keys for a specific project, using the public key
func (r *router) list(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	if project == "" {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg("db list project failed with error: no project given")
		return
	}

	AllKeys, err := r.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("db list failed with error: ", err))
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
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}

func (r *router) deleteProject(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]

	if project == "" {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg("db deleting project failed with error: no project given")
		return
	}

	AllKeys, err := r.db.list()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("db list failed with error: ", err))
		return
	}

	for _, key := range AllKeys {
		if strings.HasPrefix(key, pk+"_"+project+"_") {
			err := r.db.delete(key)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				r.logger.Error().Msg(fmt.Sprintf("db deleting key %v failed with error: %v", key, err))
				return
			}
		}
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}

// delete the value of the given key, using the public key
func (r *router) delete(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	err := r.db.delete(docKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("db deletion failed with error: ", err))
		return
	}

	w.WriteHeader(202)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is deleted successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)

}

// set the given value of the given key, using the public key
func (r *router) set(w http.ResponseWriter, request *http.Request) {

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
		r.logger.Error().Msg(fmt.Sprint("can't decode public key: ", err))
		return
	}

	// check request
	if body == "" {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg("no body given")
		return
	}

	if request.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg("no Authorization header")
		return
	}

	// verify
	verified, err := verifySignedData(body, verifyPk)
	if !verified || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("invalid data: ", err))
		return
	}
	r.logger.Debug().Msg(fmt.Sprint("signed body is verified: ", verified))

	authHeader, err := verifySignedHeader(request.Header.Get("Authorization"), verifyPk)
	if !authHeader || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("invalid authorization header: ", err))
		return
	}
	r.logger.Debug().Msg(fmt.Sprint("signed header is verified: ", err))

	// set date
	docKey := pk + "_" + projectKey
	err = r.db.set(docKey, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("database set failed with error: ", err))
		return
	}

	// response
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"msg": "data is set successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	w.Write(res)
}
