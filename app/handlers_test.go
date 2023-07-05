// Package app for pkid app
package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/pkid/client"
	"github.com/rawdaGastan/pkid/pkg"
	"github.com/stretchr/testify/assert"
)

func setUp(t testing.TB) *App {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")
	config := `{
		"port": ":3000",
		"version": "v1",
		"db_file": "pkid.db"
	}`

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	app, err := NewApp(context.Background(), configPath)
	assert.NoError(t, err)

	return app
}

func TestHandlers(t *testing.T) {
	app := setUp(t)

	privateKey, publicKey, err := client.GenerateKeyPair()
	assert.NoError(t, err)

	t.Run("test set", func(t *testing.T) {
		header := map[string]interface{}{
			"intent":    "pkid.store",
			"timestamp": time.Now().Unix(),
		}

		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		assert.NoError(t, err)

		signedHeader, err := pkg.SignEncode(header, privateKey)
		assert.NoError(t, err)

		// set request
		jsonBody := []byte(signedBody)
		bodyReader := bytes.NewReader(jsonBody)

		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, bodyReader)

		req.Header.Set("Authorization", signedHeader)
		req.Header.Set("Content-Type", "application/json")

		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.set).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusCreated)
	})

	t.Run("test set failed auth", func(t *testing.T) {
		header := map[string]interface{}{
			"intent":    "pkid.store",
			"timestamp": time.Now().Unix() - 10,
		}

		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		assert.NoError(t, err)

		signedHeader, err := pkg.SignEncode(header, privateKey)
		assert.NoError(t, err)

		// set request
		jsonBody := []byte(signedBody)
		bodyReader := bytes.NewReader(jsonBody)

		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, bodyReader)

		req.Header.Set("Authorization", signedHeader)
		req.Header.Set("Content-Type", "application/json")

		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.set).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("test set no body", func(t *testing.T) {
		header := map[string]interface{}{
			"intent":    "pkid.store",
			"timestamp": time.Now().Unix(),
		}

		signedHeader, err := pkg.SignEncode(header, privateKey)
		assert.NoError(t, err)

		// set request
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, nil)

		req.Header.Set("Authorization", signedHeader)

		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.set).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("test set no auth", func(t *testing.T) {
		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		assert.NoError(t, err)

		// set request
		jsonBody := []byte(signedBody)
		bodyReader := bytes.NewReader(jsonBody)

		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, bodyReader)

		req.Header.Set("Content-Type", "application/json")

		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.set).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("test set wrong public key", func(t *testing.T) {
		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		assert.NoError(t, err)

		// set request
		jsonBody := []byte(signedBody)
		bodyReader := bytes.NewReader(jsonBody)

		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString([]byte{}), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, bodyReader)

		req.Header.Set("Content-Type", "application/json")

		req = mux.SetURLVars(req, map[string]string{
			"pk":      "",
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.set).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("test get", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.get).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("test get empty", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.get).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("test get empty server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.list).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("test list", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "pkid")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.list).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("test list empty project", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.list).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("test delete", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.delete).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusNoContent)
	})

	t.Run("test delete empty", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.deleteProject).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusNoContent)
	})

	t.Run("test delete project", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "pkid")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.deleteProject).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusNoContent)
	})

	t.Run("test delete empty project", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "",
		})

		response := httptest.NewRecorder()
		WrapFunc(app.deleteProject).ServeHTTP(response, req)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}
