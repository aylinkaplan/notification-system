package storage

import (
	"embed"
	"io/fs"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrations embed.FS

// migrationFiles are run in order (only .up.sql)
var migrationFiles = []string{
	"migrations/001_create_notifications.up.sql",
	"migrations/002_add_source_column.up.sql",
}

func Migrate(db *sqlx.DB) error {
	for _, name := range migrationFiles {
		sql, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return err
		}
	}
	return nil
}
