package internal

import (
	"database/sql"
	"errors"
	"os"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrNotExists    = errors.New("row not exist")
	ErrSetFailed    = errors.New("set failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type sqliteDB struct {
	db *sql.DB
}

// new instance of sqlite database
func newSQLiteDB() *sqliteDB {
	return &sqliteDB{}
}

// set the connection and filePath of the sqlite db
func (sqlite *sqliteDB) setConn(filePath string) error {
	if filePath == "" {
		return errors.New("no file path provided")
	}

	os.Remove(filePath)

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return err
	}
	sqlite.db = db
	return nil
}

// create a new table pkid includes 2 columns for key and value, key is unique
func (sqlite *sqliteDB) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS pkid(
        key TEXT NOT NULL UNIQUE,
        value TEXT NOT NULL
    );
    `
	_, err := sqlite.db.Exec(query)
	return err
}

// add a new row in the table pkid with key and value
func (sqlite *sqliteDB) set(key string, value string) error {
	if key == "" {
		return errors.New("invalid key")
	}

	res, err := sqlite.db.Exec("INSERT INTO pkid(key, value) values(?,?)", key, value)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return sqlite.update(key, value)
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

// get the value of the given key in the table pkid
func (sqlite *sqliteDB) get(key string) (string, error) {
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

// update a row in the table pkid with key and value
func (sqlite *sqliteDB) update(key string, value string) error {
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

// delete the value of the given key in the table pkid
func (sqlite *sqliteDB) delete(key string) error {
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

// get all keys in the table pkid
func (sqlite *sqliteDB) list() ([]string, error) {
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
