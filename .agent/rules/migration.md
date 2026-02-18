# Database Migration Rules

## File Location

Semua migration file di `db/migrations/`

## Naming Convention

```
{timestamp}_{description}.up.sql     # Apply migration
{timestamp}_{description}.down.sql   # Rollback migration
{timestamp}_{description}_pg.up.sql     # PostgreSQL specific
{timestamp}_{description}_pg.down.sql   # PostgreSQL specific
```

- Timestamp: format `YYYYMMDDHHmmss` (contoh: `20231030144428`)
- Description: `snake_case` — `create_table_users`, `add_column_email`

## Create Table Pattern (MySQL)

```sql
CREATE TABLE {table_name}
(
    id          VARCHAR(100) NOT NULL,
    -- ... kolom lain
    created_at  BIGINT       NOT NULL,
    updated_at  BIGINT       NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY fk_{table}_{fk_column} ({fk_column}) REFERENCES {parent_table} (id)
) ENGINE = InnoDB;
```

## Create Table Pattern (PostgreSQL)

```sql
CREATE TABLE {table_name}
(
    id          VARCHAR(100) NOT NULL,
    -- ... kolom lain
    created_at  BIGINT       NOT NULL,
    updated_at  BIGINT       NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_{table}_{fk_column} FOREIGN KEY ({fk_column}) REFERENCES {parent_table} (id)
);
```

## Drop Table Pattern

```sql
DROP TABLE IF EXISTS {table_name};
```

## Rules

1. **Selalu buat up & down** — setiap migration harus punya pasangan
2. **Buat versi MySQL dan PostgreSQL** jika project support keduanya
3. **Primary key**: selalu `VARCHAR(100) NOT NULL` untuk string ID
4. **Timestamps**: `BIGINT NOT NULL` untuk `created_at` dan `updated_at` (epoch milli)
5. **Foreign key naming**: `fk_{table}_{column}` — contoh: `fk_contacts_user_id`
6. **Engine**: MySQL pakai `ENGINE = InnoDB`
7. **Nullable**: explicit — nullable tanpa `NOT NULL`, required dengan `NOT NULL`
8. **Migration tool**: gunakan `golang-migrate` atau tool sejenis
9. **Urutan**: tabel parent dibuat sebelum child (users → contacts → addresses)

## SQL Type Mapping

| Go Entity Type | MySQL        | PostgreSQL         |
| -------------- | ------------ | ------------------ |
| `string`       | `VARCHAR(N)` | `VARCHAR(N)`       |
| `int64` (time) | `BIGINT`     | `BIGINT`           |
| `int`          | `INT`        | `INTEGER`          |
| `float64`      | `DOUBLE`     | `DOUBLE PRECISION` |
| `bool`         | `TINYINT(1)` | `BOOLEAN`          |
| `text`         | `TEXT`       | `TEXT`             |
