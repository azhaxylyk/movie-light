package sql

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", "./project.db")
	if err != nil {
		return nil, err
	}

	err = CreateTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	sqlFiles := []string{
		"./migrations/users.sql",
		"./migrations/movies.sql",
		"./migrations/reviews.sql",
		"./migrations/favorites.sql",
	}

	for _, file := range sqlFiles {
		query, err := LoadSQLFile(file)
		if err != nil {
			return err
		}

		_, err = db.Exec(query)
		if err != nil {
			log.Fatalf("Ошибка выполнения SQL из файла %s: %v", file, err)
		}
	}

	return nil
}

func LoadSQLFile(filePath string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	return string(content), nil
}
