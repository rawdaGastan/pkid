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
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("error generating keys: %q", err)
	}

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

func TestPkidClientFuncsExceptions(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("error generating keys: %q", err)
	}

	t.Run("test_wrong_response_set_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode("")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Set("pkid", "key", "value", true)
		if err == nil {
			t.Error("set should fail, wrong response")
		}
	})

	t.Run("test_no_url_set_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "", 5*time.Second)
		err := c.Set("pkid", "key", "value", true)
		if err == nil {
			t.Error("set should fail, no server url")
		}
	})

	t.Run("test_wrong_url_length_set_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "postgres://user:ab", 5*time.Second)
		err := c.Set("pkid", "key", "value", true)
		if err == nil {
			t.Error("set should fail, wrong response url schema")
		}
	})

	t.Run("test_wrong_response_length_set_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Set("pkid", "key", "value", true)
		if err == nil {
			t.Error("set should fail, wrong response body length")
		}
	})

	t.Run("test_wrong_response_get_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode("")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		_, err := c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, wrong response")
		}
	})

	t.Run("test_no_url_get_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "", 5*time.Second)
		_, err := c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, no server url")
		}
	})

	t.Run("test_wrong_url_length_get_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "postgres://user:ab", 5*time.Second)
		_, err := c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, wrong response url schema")
		}
	})

	t.Run("test_wrong_response_length_get_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		_, err := c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, wrong response body length")
		}
	})

	t.Run("test_no_pub_get_func", func(t *testing.T) {
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
		c.publicKey = []byte{}
		_, err = c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, wrong public key")
		}
	})

	t.Run("test_wrong_priv_get_func", func(t *testing.T) {
		want := "value"
		payload := map[string]interface{}{
			"is_encrypted": true,
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
		c.privateKey = []byte{}
		_, err = c.Get("pkid", "key")
		if err == nil {
			t.Error("get should fail, wrong private key")
		}
	})

	t.Run("test_no_url_list_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "", 5*time.Second)
		_, err := c.List("pkid")
		if err == nil {
			t.Error("list should fail, no server url")
		}
	})

	t.Run("test_wrong_url_length_list_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "postgres://user:ab", 5*time.Second)
		_, err := c.List("pkid")
		if err == nil {
			t.Error("list should fail, wrong response url schema")
		}
	})

	t.Run("test_wrong_response_length_list_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		_, err := c.List("pkid")
		if err == nil {
			t.Error("list should fail, wrong response body length")
		}
	})

	t.Run("test_wrong_data_list_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"msg": "data is got successfully", "data": make(chan int)})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		_, err := c.List("pkid")
		if err == nil {
			t.Error("list should fail, wrong response body data")
		}
	})

	t.Run("test_no_url_delete_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "", 5*time.Second)
		err := c.Delete("pkid", "key")
		if err == nil {
			t.Error("delete should fail, no server url")
		}
	})

	t.Run("test_wrong_url_length_delete_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "postgres://user:ab", 5*time.Second)
		err := c.Delete("pkid", "key")
		if err == nil {
			t.Error("delete should fail, wrong response url schema")
		}
	})

	t.Run("test_wrong_response_length_delete_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Delete("pkid", "key")
		if err == nil {
			t.Error("delete should fail, wrong response body length")
		}
	})

	t.Run("test_unmarshal_res_delete_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode("")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.Delete("pkid", "key")
		if err == nil {
			t.Error("delete should fail, wrong unmarshal response")
		}
	})

	t.Run("test_no_url_project_delete_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "", 5*time.Second)
		err := c.DeleteProject("pkid")
		if err == nil {
			t.Error("delete project should fail, no server url")
		}
	})

	t.Run("test_wrong_url_length_delete_project_func", func(t *testing.T) {
		c := NewPkidClient(privateKey, publicKey, "postgres://user:ab", 5*time.Second)
		err := c.DeleteProject("pkid")
		if err == nil {
			t.Error("delete project should fail, wrong response url schema")
		}
	})

	t.Run("test_wrong_response_length_delete__project_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.DeleteProject("pkid")
		if err == nil {
			t.Error("delete project should fail, wrong response body length")
		}
	})

	t.Run("test_unmarshal_res_delete_project_func", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode("")
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.DeleteProject("pkid")
		if err == nil {
			t.Error("delete project should fail, wrong unmarshal response")
		}
	})

}

func TestPkidClientFuncs(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("error generating keys: %q", err)
	}

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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"msg": "data is got successfully", "data": want})
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"msg": "data is deleted successfully"})
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"msg": "data is deleted successfully"})
		}))

		c := NewPkidClient(privateKey, publicKey, s.URL, 5*time.Second)
		err := c.DeleteProject("pkid")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPkidKeys(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("error generating keys: %q", err)
	}

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

	t.Run("test_corrupt_seed_key_pair", func(t *testing.T) {
		seed := "XXXXXaGVsbG8="
		_, _, err := GenerateKeyPairUsingSeed(seed)
		if err == nil {
			t.Error("generating keys should fail")
		}
	})
}

func TestPkidClient(t *testing.T) {
	url := "http://localhost:3000"

	// check server running
	_, err := http.Get(url)
	if err == nil {
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Errorf("error generating keys: %q", err)
		}
		pkidClient := NewPkidClient(privateKey, publicKey, url, 5*time.Second)

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
