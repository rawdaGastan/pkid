package internal

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
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

func TestVerifiers(t *testing.T) {
	_, publicKey, _ := sodium.CryptoSignKeyPair()

	t.Run("test_wrong_encoding", func(t *testing.T) {
		encoded := "XXXXXaGVsbG8="

		_, err := verifySignedData(encoded, publicKey)
		if err == nil {
			t.Error("decoding should fail")
		}
	})

	t.Run("test_wrong_encoding_header", func(t *testing.T) {
		encoded := "XXXXXaGVsbG8="

		_, err := verifySignedHeader(encoded, publicKey)
		if err == nil {
			t.Error("decoding should fail")
		}
	})
}

func TestServer(t *testing.T) {
	testDir := t.TempDir()
	pkidStore := NewSqliteStore()

	logger := zerolog.New(os.Stdout).With().Logger()
	privateKey, publicKey, _ := sodium.CryptoSignKeyPair()

	router := newRouter(logger, pkidStore)
	err := router.setConn(testDir + "/pkid.db")

	if err != nil {
		t.Errorf(fmt.Sprint("error starting server database: ", err))
	}

	t.Run("test_failed_server", func(t *testing.T) {
		_, err := NewServer(logger, []mux.MiddlewareFunc{}, pkidStore, "", 3000)

		if err == nil {
			t.Errorf("expected error got nil")
		}

	})

	t.Run("test_success_server", func(t *testing.T) {
		mws := []mux.MiddlewareFunc{}
		loggingMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println(r.RequestURI)
				next.ServeHTTP(w, r)
			})
		}
		mws = append(mws, loggingMw)
		_, err := NewServer(logger, mws, pkidStore, "pkid.db", 3000)

		fmt.Printf("err: %v\n", err)

		if err != nil {
			t.Errorf("server should be created")
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
		router.set(w, req)
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
		router.set(w, req)
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
		router.set(w, req)
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
		router.set(w, req)
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
		router.set(w, req)
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
		router.get(w, req)
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
		router.get(w, req)
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
		router.list(w, req)
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
		router.list(w, req)
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
		router.delete(w, req)
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
		router.delete(w, req)
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
		router.deleteProject(w, req)
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
		router.deleteProject(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode == 202 {
			t.Errorf("delete should fail")
		}
	})

}
