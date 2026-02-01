# Database Migrations with golang-migrate

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

## Quick Start

### 1. Create a migration

```bash
make migrate-create name=add_user_avatar
```

Creates: `migrations/20250125120000_add_user_avatar.up.sql` and `.down.sql`

### 2. Run migrations

```bash
make migrate-up
```

Runs all pending migrations in order.

### 3. Check status

```bash
make migrate-status
```

Shows current migration version.

### 4. Rollback

```bash
make migrate-down
```

Rolls back last migration.

### 5. Reset database

```bash
make db-reset
```

Drops schema and re-runs all migrations.

## Migration Files

Migrations are in `./migrations/` directory with format:
```
YYYYMMDDHHMMSS_description.up.sql
YYYYMMDDHHMMSS_description.down.sql
```

### Example Migration

**Up migration** (`20250125120000_add_users_table.up.sql`):
```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()) * 1000)
);
```

**Down migration** (`20250125120000_add_users_table.down.sql`):
```sql
-- +migrate Down
DROP TABLE IF EXISTS users;
```

## Make Targets

| Target | Description |
|--------|-------------|
| `migrate-up` | Run all pending migrations |
| `migrate-down` | Rollback last migration |
| `migrate-create name=<name>` | Create new migration |
| `migrate-status` | Show migration status |
| `db-reset` | Drop schema and re-run migrations |

## How It Works

golang-migrate automatically:
- Creates `schema_migrations` table to track applied migrations
- Runs migrations in filename order (chronological)
- Supports rollback via `.down.sql` files
- Handles migration errors gracefully (dirty state)
- Supports multiple database drivers (PostgreSQL, MySQL, etc.)

## Best Practices

### 1. Always create both .up and .down files

```bash
make migrate-create name=add_column
# Creates:
# - 20250125120000_add_column.up.sql
# - 20250125120000_add_column.down.sql
```

### 2. Use idempotent SQL

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (...);
DROP INDEX IF EXISTS idx_users_email;

-- +migrate Down
DROP TABLE IF EXISTS users CASCADE;
```

### 3. One logical change per migration

- ✅ Good: `add_user_avatar.up.sql`
- ❌ Bad: `add_avatar_and_rename_table.up.sql`

### 4. Test migrations locally

```bash
# Create test database
export TEST_DB_URL="postgres://localhost:5432/test_db"

# Run migrations
DATABASE_URL=$TEST_DB_URL make migrate-up

# Verify
psql $TEST_DB_URL -c "\d"
```

### 5. Version control your migrations

Migrations are part of your application history. Never modify a migration after it's been applied to production. Instead:
1. Keep the old migration
2. Create a new migration to fix/undo the change

## Troubleshooting

### Migration failed?

```bash
# Check what went wrong
make migrate-status

# Force reset (use with caution!)
make db-reset
```

### Database is in "dirty" state?

```bash
# This means a migration failed mid-execution
# Check logs to see the error
# Fix the migration file
# Run: migrate up --force (via CLI)
```

### Need to re-run a migration?

golang-migrate tracks applied migrations. To re-run:

**Option 1:** Rollback and re-run
```bash
make migrate-down
make migrate-up
```

**Option 2:** Force specific version (advanced)
```bash
# Use golang-migrate CLI directly
migrate -database "$DATABASE_URL" -path ./migrations force <version>
```

## Examples

### Adding a column
```sql
-- +migrate Up
ALTER TABLE users ADD COLUMN avatar_url TEXT;

-- +migrate Down
ALTER TABLE users DROP COLUMN avatar_url;
```

### Creating a table
```sql
-- +migrate Up
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE posts CASCADE;
```

### Renaming a table
```sql
-- +migrate Up
ALTER TABLE users RENAME TO profiles;

-- +migrate Down
ALTER TABLE profiles RENAME TO users;
```

### Creating an index
```sql
-- +migrate Up
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_email;
```

## Advanced Usage

### Manual CLI usage

```bash
# Run golang-migrate directly with go run
go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  -database "$DATABASE_URL" \
  -path ./migrations \
  up

# Rollback specific number of migrations
go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  -database "$DATABASE_URL" \
  -path ./migrations \
  down 2

# Force migration version
go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  -database "$DATABASE_URL" \
  -path ./migrations \
  force 20250125120000
```

### Database URL format

**PostgreSQL:**
```
postgres://user:password@localhost:5432/dbname?sslmode=disable
```

**With connection pool:**
```
postgres://user:password@localhost:5432/dbname?sslmode=disable&pool_max_conns=10
```

## Related Files

- `./migrations/` - Migration SQL files
- `Makefile` - Migration make targets
- `go.mod` - golang-migrate dependency

## Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [Migration Guide](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md)
- [PostgreSQL Best Practices](https://wiki.postgresql.org/wiki/Don%27t-Do-This)

## See Also

- [OBSERVABILITY.md](../OBSERVABILITY.md) - Debugging guide
- [README.md](../README.md) - Project overview
