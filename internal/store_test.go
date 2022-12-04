package internal

import "testing"

func TestPkidStore(t *testing.T) {
	testDir := t.TempDir()
	pkidStore := NewSqliteStore()

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
