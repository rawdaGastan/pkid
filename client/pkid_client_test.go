package client

import (
	"testing"
	"time"
)

func TestPkidClient(t *testing.T) {
	url := "http://localhost:3000"
	privateKey, publicKey := GenerateKeyPair()
	pkidClient := NewPkidClient(privateKey, publicKey, url, 5*time.Second)

	t.Run("test_seed_key_pair", func(t *testing.T) {
		seed := "bm2xl92552zz0Kxtvg4Gbaosnh6FY9H2WsIKao6Emh8="
		_, _, err := GenerateKeyPairUsingSeed(seed)
		if err != nil {
			t.Errorf("generating keys should be successful: %v", err)
		}
	})

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
