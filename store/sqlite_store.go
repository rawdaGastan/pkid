// package store is for pkid storage
package store

import (
	"database/sql"
	"errors"

	sqlite3 "github.com/mattn/go-sqlite3"
)

var (
	// ErrNotExists is an error for non existing rows in db
	ErrNotExists = errors.New("row not exist")
	// ErrSetFailed is an error when setting data fails
	ErrSetFailed = errors.New("set failed")
	// ErrDeleteFailed is an error when deleting data fails
	ErrDeleteFailed = errors.New("deletion failed")
)

// SqliteStore is a struct for sqlite store requirements
type SqliteStore struct {
	db *sql.DB
}

// NewSqliteStore creates a new instance of sqlite database
func NewSqliteStore() *SqliteStore {
	return &SqliteStore{}
}

// SetConn sets the connection and filePath of the sqlite db
func (sqlite *SqliteStore) SetConn(filePath string) error {
	if filePath == "" {
		return errors.New("no file path provided")
	}

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return err
	}
	sqlite.db = db
	return nil
}

// Migrate creates a new table includes 2 columns for key and value, key is unique
func (sqlite *SqliteStore) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS pkid(
        key TEXT NOT NULL UNIQUE,
        value TEXT NOT NULL
    );
    `
	_, err := sqlite.db.Exec(query)
	return err
}

// Set adds a new row with key and value
func (sqlite *SqliteStore) Set(key string, value string) error {
	if key == "" {
		return errors.New("invalid key")
	}

	res, err := sqlite.db.Exec("INSERT INTO pkid(key, value) values(?,?)", key, value)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return sqlite.Update(key, value)
			}
		}
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSetFailed
	}

	return nil
}

// Get gets the value of the given key
func (sqlite *SqliteStore) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("invalid key")
	}

	row := sqlite.db.QueryRow("SELECT * FROM pkid WHERE key = ?", key)

	var value string
	if err := row.Scan(&key, &value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotExists
		}
		return "", err
	}
	return value, nil
}

// Update updates a row with key and value
func (sqlite *SqliteStore) Update(key string, value string) error {
	if key == "" {
		return errors.New("invalid updated ID")
	}
	res, err := sqlite.db.Exec("UPDATE pkid SET value = ?", value)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSetFailed
	}

	return nil
}

// Delete deletes the value of the given key
func (sqlite *SqliteStore) Delete(key string) error {
	if key == "" {
		return errors.New("invalid key")
	}

	res, err := sqlite.db.Exec("DELETE FROM pkid WHERE key = ?", key)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}

// List gets all keys
func (sqlite *SqliteStore) List() ([]string, error) {
	rows, err := sqlite.db.Query("SELECT * FROM pkid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []string
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		all = append(all, key)
	}
	return all, nil
}
