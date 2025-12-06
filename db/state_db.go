package db

import (
	"database/sql"
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/duckdb/duckdb-go/v2"
)

type StateDB struct {
	DB *sql.DB
}

func NewDB() *StateDB {
	db, err := sql.Open("duckdb", "icebreaker.db")
	if err != nil {
		log.Error("Failed to open database: %v", err)
		os.Exit(1)
	}

	return &StateDB{DB: db}
}

func (s *StateDB) Close() error {
	return s.DB.Close()
}

func (s *StateDB) Migrate() error {
	_, err := s.DB.Exec("CREATE TABLE IF NOT EXISTS keys (key TEXT PRIMARY KEY)")
	if err != nil {
		return err
	}

	return nil
}

func (s *StateDB) AddKey(key string) error {
	_, err := s.DB.Exec("INSERT OR IGNORE INTO keys (key) VALUES (?)", key)
	if err != nil {
		return err
	}
	return nil
}

func (s *StateDB) GetKeys() ([]string, error) {
	rows, err := s.DB.Query("SELECT key FROM keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var key string
		err = rows.Scan(&key)

		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *StateDB) DeleteKey(key string) error {
	_, err := s.DB.Exec("DELETE FROM keys WHERE key = ?", key)
	if err != nil {
		return err
	}
	return nil
}

// TruncateKeys removes all keys from the keys table.
func (s *StateDB) TruncateKeys() error {
	_, err := s.DB.Exec("DELETE FROM keys")
	if err != nil {
		return err
	}
	return nil
}
