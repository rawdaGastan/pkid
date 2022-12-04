package internal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
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
func newRouter(logger zerolog.Logger, db PkidStore) router {
	return router{
		db:     db,
		logger: logger,
	}
}

// set the connection and migration of the db
func (r *router) setConn(filePath string) error {
	err := r.db.setConn(filePath)
	if err != nil {
		return err
	}

	if err := r.db.migrate(); err != nil {
		return err
	}
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

	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("write response failed: ", err))
		return
	}
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
	res, err := json.Marshal(map[string]interface{}{"data": keys, "msg": "data is got successfully"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("response failed with error: ", err))
		return
	}
	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("write response failed: ", err))
		return
	}
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
	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("write response failed: ", err))
		return
	}
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
	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("write response failed: ", err))
		return
	}

}

// set the given value of the given key, using the public key
func (r *router) set(w http.ResponseWriter, request *http.Request) {

	pk := mux.Vars(request)["pk"]
	project := mux.Vars(request)["project"]
	key := mux.Vars(request)["key"]
	projectKey := project + "_" + key

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(request.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("can't read from body buffer: ", err))
		return
	}

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
	r.logger.Debug().Msg(fmt.Sprint("signed header is verified: ", authHeader))

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
	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		r.logger.Error().Msg(fmt.Sprint("write response failed: ", err))
		return
	}
}
