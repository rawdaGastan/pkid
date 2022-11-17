package internal

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	sodium "github.com/gokillers/libsodium-go/cryptosign"
	"github.com/gorilla/mux"
	"github.com/rawdaGastan/pkid/pkg"
	"github.com/rs/zerolog"
)

func TestPkidStore(t *testing.T) {
	testDir := t.TempDir()
	pkidStore := newPkidStore()

	t.Run("test_empty_file", func(t *testing.T) {
		err := pkidStore.setConn("")

		if err == nil {
			t.Errorf("connection should not be set")
		}
	})

	t.Run("test_connection", func(t *testing.T) {
		err := pkidStore.setConn(testDir + "/pkid.db")

		if err != nil {
			t.Errorf("connection should be set")
		}
	})

	t.Run("test_migrate", func(t *testing.T) {
		err := pkidStore.migrate()
		if err != nil {
			t.Errorf("migration should succeed")
		}
	})

	t.Run("test_set", func(t *testing.T) {
		err := pkidStore.set("key", "value")
		if err != nil {
			t.Errorf("set should succeed")
		}
	})

	t.Run("test_set_update", func(t *testing.T) {
		err := pkidStore.set("key", "valueUpdated")
		if err != nil {
			t.Errorf("set should succeed")
		}
	})

	t.Run("test_get", func(t *testing.T) {
		value, err := pkidStore.get("key")
		if err != nil {
			t.Errorf("get should not fail: %v", err)
		}

		if value != "valueUpdated" {
			t.Errorf("value of the key should be value")
		}
	})

	t.Run("test_list", func(t *testing.T) {
		keys, err := pkidStore.list()
		if err != nil {
			t.Errorf("list should not fail: %v", err)
		}

		if len(keys) != 1 {
			t.Errorf("keys should include one key")
		}
	})

	t.Run("test_delete", func(t *testing.T) {
		err := pkidStore.delete("key")
		if err != nil {
			t.Errorf("delete should not fail: %v", err)
		}
	})

	t.Run("test_get_deleted", func(t *testing.T) {
		_, err := pkidStore.get("key")
		if err == nil {
			t.Errorf("get should fail")
		}
	})

	t.Run("test_delete_deleted", func(t *testing.T) {
		err := pkidStore.delete("key")
		if err == nil {
			t.Errorf("delete should fail")
		}
	})

	t.Run("test_list_empty", func(t *testing.T) {
		keys, err := pkidStore.list()
		if err != nil {
			t.Errorf("list should not fail: %v", err)
		}

		if len(keys) > 0 {
			t.Errorf("list should be empty")
		}
	})

	t.Run("test_set_empty", func(t *testing.T) {
		err := pkidStore.set("", "value")
		if err == nil {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_set_update_empty", func(t *testing.T) {
		err := pkidStore.set("", "valueUpdated")
		if err == nil {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_get_empty", func(t *testing.T) {
		_, err := pkidStore.get("")
		if err == nil {
			t.Errorf("get should fail")
		}
	})

	t.Run("test_delete_empty", func(t *testing.T) {
		err := pkidStore.delete("")
		if err == nil {
			t.Errorf("delete should fail")
		}
	})

	t.Run("test_update_empty", func(t *testing.T) {
		err := pkidStore.update("", "value")
		if err == nil {
			t.Errorf("update should fail")
		}
	})

	t.Run("test_update_empty", func(t *testing.T) {
		err := pkidStore.update("key", "value")
		if err == nil {
			t.Errorf("update should fail")
		}
	})
}

func TestServer(t *testing.T) {
	testDir := t.TempDir()

	logger := zerolog.New(os.Stdout).With().Logger()
	privateKey, publicKey, _ := sodium.CryptoSignKeyPair()

	server := newServer(logger)
	err := server.setConn(testDir + "/pkid.db")

	if err != nil {
		t.Errorf(fmt.Sprint("error starting server database: ", err))
	}

	t.Run("test_failed_server", func(t *testing.T) {
		err := StartServer(logger, "", 3000)

		if err == nil {
			t.Errorf("expected error got nil")
		}

	})

	t.Run("test_set_server", func(t *testing.T) {
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
		if err != nil {
			t.Errorf("error sign body: %v", err)
		}

		signedHeader, err := pkg.SignEncode(header, privateKey)
		if err != nil {
			t.Errorf("error sign header: %v", err)
		}

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

		w := httptest.NewRecorder()
		server.set(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != 201 {
			t.Errorf("set should be successful")
		}
	})

	t.Run("test_set_failed_auth_server", func(t *testing.T) {
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
		if err != nil {
			t.Errorf("error sign body: %v", err)
		}

		signedHeader, err := pkg.SignEncode(header, privateKey)
		if err != nil {
			t.Errorf("error sign header: %v", err)
		}

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

		w := httptest.NewRecorder()
		server.set(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 201 {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_set_no_body", func(t *testing.T) {
		header := map[string]interface{}{
			"intent":    "pkid.store",
			"timestamp": time.Now().Unix(),
		}

		signedHeader, err := pkg.SignEncode(header, privateKey)
		if err != nil {
			t.Errorf("error sign header: %v", err)
		}

		// set request
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodPost, requestURL, nil)

		req.Header.Set("Authorization", signedHeader)

		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})

		w := httptest.NewRecorder()
		server.set(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 201 {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_set_no_auth", func(t *testing.T) {

		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		if err != nil {
			t.Errorf("error sign body: %v", err)
		}

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

		w := httptest.NewRecorder()
		server.set(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 201 {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_set_wrong_key", func(t *testing.T) {

		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      "value",
			"data_version": 1,
		}

		signedBody, err := pkg.SignEncode(payload, privateKey)
		if err != nil {
			t.Errorf("error sign body: %v", err)
		}

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

		w := httptest.NewRecorder()
		server.set(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 201 {
			t.Errorf("set should fail")
		}
	})

	t.Run("test_get_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})
		server.get(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("get should be successful")
		}
	})

	t.Run("test_get_empty_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "",
		})
		server.get(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 200 {
			t.Errorf("get should fail")
		}
	})

	t.Run("test_list_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "pkid")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
		})
		server.list(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("list should be successful")
		}
	})

	t.Run("test_list_empty_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "")
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "",
		})
		server.list(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 200 {
			t.Errorf("list should fail")
		}
	})

	t.Run("test_delete_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "key")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "key",
		})
		server.delete(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != 202 {
			t.Errorf("delete should be successful")
		}
	})

	t.Run("test_delete_empty_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", "")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
			"key":     "",
		})
		server.delete(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 202 {
			t.Errorf("delete should fail")
		}
	})

	t.Run("test_delete_project_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "pkid")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "pkid",
		})
		server.deleteProject(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != 202 {
			t.Errorf("delete should be successful")
		}
	})

	t.Run("test_delete_project_empty_server", func(t *testing.T) {
		requestURL := fmt.Sprintf("/%v/%v", hex.EncodeToString(publicKey), "")
		req := httptest.NewRequest(http.MethodDelete, requestURL, nil)
		w := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{
			"pk":      hex.EncodeToString(publicKey),
			"project": "",
		})
		server.deleteProject(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 202 {
			t.Errorf("delete should fail")
		}
	})

}
