package migration

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/artufi/trader/infrastructure/database/migration/migrations"
	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB) ([]string, error) {
	goose.SetBaseFS(migrations.FSMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("goose up migration: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return nil, fmt.Errorf("goose up migration: %w", err)
	}

	return getAllFilenames(migrations.FSMigrations)
}

func Down(db *sql.DB) ([]string, error) {
	goose.SetBaseFS(migrations.FSMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("goose down migration: %w", err)
	}
	if err := goose.Down(db, "."); err != nil {
		return nil, fmt.Errorf("goose down migration: %w", err)
	}

	return getAllFilenames(migrations.FSMigrations)
}

func getAllFilenames(efsys fs.FS) ([]string, error) {
	var filenames []string
	err := fs.WalkDir(efsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		filenames = append(filenames, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("goose migration: %w", err)
	}

	return filenames, nil
}
