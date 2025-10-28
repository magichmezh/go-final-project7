package db

import (
	"database/sql"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255),
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := false
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			install = true
		} else {
			return err
		}
	}

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		_, err = DB.Exec(schema)
		if err != nil {
			return err
		}
	}
	return nil
}
