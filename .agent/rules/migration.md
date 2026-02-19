# Database Migration Rules

## File Location

Semua migration file di `db/migrations/`

## Naming Convention

```
{timestamp}_{description}.up.sql     # Apply migration
{timestamp}_{description}.down.sql   # Rollback migration
```

- Timestamp: format `YYYYMMDDHHmmss` (contoh: `20231030144428`)
- Description: `snake_case` — `create_initial_schema`, `add_column_email`

## Create Table Pattern (PostgreSQL)

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE {table_name}
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- ... kolom lain
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_{child}_{parent} FOREIGN KEY ({fk_column}) REFERENCES {parent_table} (id) ON DELETE CASCADE
);

CREATE INDEX idx_{table}_{column} ON {table_name} ({column});
```

## Drop Table Pattern

```sql
DROP TABLE IF EXISTS {table_name};
DROP EXTENSION IF EXISTS pgcrypto;
```

## Rules

1. **Selalu buat up & down** — setiap migration harus punya pasangan
2. **Project ini PostgreSQL only** — tidak perlu buat versi MySQL
3. **Primary key**: `UUID PRIMARY KEY DEFAULT gen_random_uuid()` — gunakan `pgcrypto` extension
4. **Timestamps**: `TIMESTAMPTZ NOT NULL DEFAULT now()` untuk `created_at` dan `updated_at`
5. **Foreign key naming**: `fk_{child_table}_{parent_table}` — contoh: `fk_refresh_tokens_users`
6. **Nullable**: explicit — nullable tanpa `NOT NULL`, required dengan `NOT NULL`
7. **Migration tool**: gunakan `golang-migrate` — `migrate -database "postgres://..." -path db/migrations up`
8. **Urutan**: tabel parent dibuat sebelum child (users → refresh_tokens → ...)
9. **Down migration**: drop tabel dalam urutan terbalik (child dulu, lalu parent)

## SQL Type Mapping

| Go Entity Type | PostgreSQL         | Keterangan                   |
| -------------- | ------------------ | ---------------------------- |
| `string`       | `VARCHAR(N)`       | Text, string ID              |
| `time.Time`    | `TIMESTAMPTZ`      | Timestamps (created/updated) |
| `*time.Time`   | `TIMESTAMPTZ`      | Nullable timestamps          |
| `int`          | `INTEGER`          | Angka umum                   |
| `*int`         | `INTEGER`          | Nullable integer             |
| `float64`      | `DOUBLE PRECISION` | Angka desimal                |
| `bool`         | `BOOLEAN`          | Flag                         |
| `text`         | `TEXT`             | Long text                    |
