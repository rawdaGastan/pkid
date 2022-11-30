package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rawdaGastan/pkid/pkg"
)

type FakePkid func(*http.Request) (*http.Response, error)

func (f FakePkid) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFakePkid(t *testing.T) {
	privateKey, publicKey := GenerateKeyPair()

	t.Run("test_set_func", func(t *testing.T) {
		client := &http.Client{
			Transport: FakePkid(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"msg": "data is set successfully"}`)),
				}, nil
			}),
		}

		cl := NewPkidClientWithHTTPClient(privateKey, publicKey, "", client)

		err := cl.Set("pkid", "key", "value", false)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test_get_func", func(t *testing.T) {
		want := "value"
		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      want,
			"data_version": 1,
		}
		signedBody, err := pkg.SignEncode(payload, privateKey)
		if err != nil {
			t.Fatal(err)
		}

		client := &http.Client{
			Transport: FakePkid(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"msg": "data is got successfully", "data": "` + signedBody + `"}`)),
				}, nil
			}),
		}

		cl := NewPkidClientWithHTTPClient(privateKey, publicKey, "", client)

		got, err := cl.Get("pkid", "key")
		if err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Errorf("Unexpected pkid returned. Want %q, got %q", want, got)
		}
	})

	t.Run("test_list_func", func(t *testing.T) {
		want := []string{}
		client := &http.Client{
			Transport: FakePkid(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"msg": "data is got successfully", "data": ["key"]}`)),
				}, nil
			}),
		}

		cl := NewPkidClientWithHTTPClient(privateKey, publicKey, "", client)

		got, err := cl.List("pkid")
		if err != nil {
			t.Fatal(err)
		}

		if reflect.DeepEqual(got, want) {
			t.Errorf("Unexpected pkid returned. Want %q, got %q", want, got)
		}
	})

	t.Run("test_delete_func", func(t *testing.T) {
		client := &http.Client{
			Transport: FakePkid(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"msg": "data is deleted successfully"}`)),
				}, nil
			}),
		}

		cl := NewPkidClientWithHTTPClient(privateKey, publicKey, "", client)

		err := cl.Delete("pkid", "key")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test_delete_proj_func", func(t *testing.T) {
		client := &http.Client{
			Transport: FakePkid(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"msg": "data is deleted successfully"}`)),
				}, nil
			}),
		}

		cl := NewPkidClientWithHTTPClient(privateKey, publicKey, "", client)

		err := cl.DeleteProject("pkid")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPkidClientFuncs(t *testing.T) {
	privateKey, publicKey := GenerateKeyPair()

	t.Run("test_set_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"msg": "data is set successfully"})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Set("pkid", "key", "value", true)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test_get_func", func(t *testing.T) {
		want := "value"
		payload := map[string]interface{}{
			"is_encrypted": false,
			"payload":      want,
			"data_version": 1,
		}
		signedBody, err := pkg.SignEncode(payload, privateKey)
		if err != nil {
			t.Fatal(err)
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"msg": "data is got successfully", "data": signedBody})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		got, err := c.Get("pkid", "key")
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("Unexpected pkid returned. Got %q, want %q", got, want)
		}
	})

	t.Run("test_list_func", func(t *testing.T) {
		want := []string{"key"}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"msg": "data is got successfully", "data": want})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		got, err := c.List("pkid")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Unexpected pkid returned. Got %q, want %q", got, want)
		}
	})

	t.Run("test_delete_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"msg": "data is deleted successfully"})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Delete("pkid", "key")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test_delete_project_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"msg": "data is deleted successfully"})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.DeleteProject("pkid")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPkidClient(t *testing.T) {
	url := "http://localhost:3000"
	privateKey, publicKey := GenerateKeyPair()
	pkidClient := NewPkidClient(privateKey, publicKey, url, 5*time.Second)

	t.Run("test_pub_from_priv", func(t *testing.T) {
		testPublicKey := GetPublicKey(privateKey)

		if !reflect.DeepEqual(testPublicKey, publicKey) {
			t.Errorf("public key is wrong")
		}
	})

	t.Run("test_seed_key_pair", func(t *testing.T) {
		seed := "bm2xl92552zz0Kxtvg4Gbaosnh6FY9H2WsIKao6Emh8="
		_, _, err := GenerateKeyPairUsingSeed(seed)
		if err != nil {
			t.Errorf("generating keys should be successful: %v", err)
		}
	})

	// check server running
	_, err := http.Get(url)
	if err == nil {

		t.Run("test_set", func(t *testing.T) {
			err := pkidClient.Set("pkid", "key", "value", true)
			if err != nil {
				t.Errorf("set should be successful: %v", err)
			}
		})

		t.Run("test_get", func(t *testing.T) {
			value, err := pkidClient.Get("pkid", "key")
			if err != nil {
				t.Errorf("get should be successful: %v", err)
			}

			if value == "value" {
				t.Errorf("get should be successful")
			}
		})

		t.Run("test_list", func(t *testing.T) {
			keys, err := pkidClient.List("pkid")
			if err != nil {
				t.Errorf("list should be successful: %v", err)
			}

			if keys[0] != "key" {
				t.Errorf("list should return key value")
			}
		})

		t.Run("test_delete", func(t *testing.T) {
			err := pkidClient.Delete("pkid", "key")
			if err != nil {
				t.Errorf("delete should be successful: %v", err)
			}
		})

		t.Run("test_delete_project", func(t *testing.T) {
			err := pkidClient.DeleteProject("pkid")
			if err != nil {
				t.Errorf("delete should be successful: %v", err)
			}
		})
	}
}
