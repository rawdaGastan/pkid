// Package app for pkid app
package app

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// get the value of the given key, using the public key
func (a *App) get(r *http.Request) (interface{}, Response) {
	pk := mux.Vars(r)["pk"]
	project := mux.Vars(r)["project"]
	key := mux.Vars(r)["key"]
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	value, err := a.db.Get(docKey)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(fmt.Errorf("can't find key: %s", docKey))
	}

	return ResponseMsg{
		Message: "data is got successfully",
		Data:    value,
	}, Ok()
}

// list all keys for a specific project, using the public key
func (a *App) list(r *http.Request) (interface{}, Response) {

	pk := mux.Vars(r)["pk"]
	project := mux.Vars(r)["project"]

	if project == "" {
		return nil, BadRequest(errors.New("db list project failed with error: no project given"))
	}

	AllKeys, err := a.db.List()
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New("db list failed"))
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

	return ResponseMsg{
		Message: "data is listed successfully",
		Data:    keys,
	}, Ok()
}

func (a *App) deleteProject(r *http.Request) (interface{}, Response) {
	pk := mux.Vars(r)["pk"]
	project := mux.Vars(r)["project"]

	if project == "" {
		return nil, BadRequest(errors.New("db list project failed with error: no project given"))
	}

	AllKeys, err := a.db.List()
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New("db list failed"))
	}

	for _, key := range AllKeys {
		if strings.HasPrefix(key, pk+"_"+project+"_") {
			err := a.db.Delete(key)
			if err != nil {
				log.Error().Err(err).Send()
				return nil, InternalServerError(fmt.Errorf("db deleting key %s failed", key))
			}
		}
	}

	return ResponseMsg{
		Message: "project is deleted successfully",
		Data:    nil,
	}, Deleted()
}

// delete the value of the given key, using the public key
func (a *App) delete(r *http.Request) (interface{}, Response) {
	pk := mux.Vars(r)["pk"]
	project := mux.Vars(r)["project"]
	key := mux.Vars(r)["key"]
	projectKey := project + "_" + key

	docKey := pk + "_" + projectKey
	err := a.db.Delete(docKey)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(("db deletion failed")))
	}

	return ResponseMsg{
		Message: "data is deleted successfully",
		Data:    nil,
	}, Deleted()
}

// set the given value of the given key, using the public key
func (a *App) set(r *http.Request) (interface{}, Response) {
	pk := mux.Vars(r)["pk"]
	project := mux.Vars(r)["project"]
	key := mux.Vars(r)["key"]
	projectKey := project + "_" + key

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(("failed to read body")))
	}

	body := buf.String()

	if body == "" {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(("no body is provided")))
	}

	// verify key
	if len(pk) == 0 {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(("public key is empty")))
	}

	verifyPk, err := hex.DecodeString(pk)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(("cannot verify public key")))
	}

	if r.Header.Get("Authorization") == "" {
		log.Error().Err(err).Send()
		return nil, UnAuthorized(errors.New(("no Authorization is provided")))
	}

	// verify
	verified, err := verifySignedData(body, verifyPk)
	if !verified || err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(("invalid data")))
	}

	authHeader, err := verifySignedHeader(r.Header.Get("Authorization"), verifyPk)
	if !authHeader || err != nil {
		log.Error().Err(err).Send()
		return nil, UnAuthorized(errors.New(("invalid authorization header")))
	}

	// set date
	docKey := pk + "_" + projectKey
	err = a.db.Set(docKey, body)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(("database set failed")))
	}

	// response
	return ResponseMsg{
		Message: "data is set successfully",
		Data:    nil,
	}, Created()
}
