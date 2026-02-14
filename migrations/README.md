# Database Migrations

Migrations are embedded in the application and run automatically on startup.

Schema files: `internal/storage/migrations/`

- `001_create_notifications.up.sql` – notifications table
- `002_add_source_column.up.sql` – `source` column (for testing)

**To test:** Start the application; migrations run automatically. When adding a new migration, add its filename to the `migrationFiles` list in `migrate.go`.
