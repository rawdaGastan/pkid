package client

import (
	"testing"
)

func TestSqliteDB(t *testing.T) {
	port := 3000
	privateKey, publicKey := GenerateKeyPair()
	pkidClient := NewPkidClient(privateKey, publicKey, port)

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
}
