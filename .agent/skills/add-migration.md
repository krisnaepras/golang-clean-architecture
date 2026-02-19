# Skill: Add Database Migration

## Trigger

Gunakan skill ini ketika perlu menambah atau mengubah tabel database (create table, add column, add index, dll).

## Prerequisites

- Deskripsi perubahan database yang diinginkan
- Project ini **PostgreSQL only** — tidak perlu MySQL migration

## Steps

### Step 1: Generate Timestamp

Gunakan format `YYYYMMDDHHmmss` berdasarkan waktu saat ini.

Contoh: `20260218100000`

### Step 2: Create Up Migration

**File**: `db/migrations/{timestamp}_{description}.up.sql`

Contoh — Create table:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE {table_name}
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT         NULL,
    parent_id   UUID         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT fk_{child}_{parent} FOREIGN KEY (parent_id) REFERENCES {parent_table} (id) ON DELETE CASCADE
);

CREATE INDEX idx_{table}_{column} ON {table_name} ({column});
```

Contoh — Add column:

```sql
ALTER TABLE {table_name} ADD COLUMN {column_name} {type} {constraint};
```

Contoh — Add index:

```sql
CREATE INDEX idx_{table}_{column} ON {table_name} ({column});
```

### Step 3: Create Down Migration

**File**: `db/migrations/{timestamp}_{description}.down.sql`

Contoh — Drop table:

```sql
DROP TABLE IF EXISTS {table_name};
DROP EXTENSION IF EXISTS pgcrypto;
```

Contoh — Remove column:

```sql
ALTER TABLE {table_name} DROP COLUMN {column_name};
```

### Step 4: Run Migration

```bash
migrate -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" -path db/migrations up
```

### Step 5: Update Entity (jika perlu)

Jika migration menambah kolom baru, update struct entity di `internal/entity/{name}_entity.go`.

### Step 5: Update Models (jika perlu)

Jika ada field baru, update:

- Response model
- Request models
- Event model (jika di-publish ke Kafka)
- Converter functions

## Naming Convention

| Operasi       | Format Deskripsi                   |
| ------------- | ---------------------------------- |
| Create table  | `create_table_{name}`              |
| Add column    | `add_{column}_to_{table}`          |
| Remove column | `remove_{column}_from_{table}`     |
| Add index     | `add_index_{column}_on_{table}`    |
| Add FK        | `add_fk_{table}_{column}`          |
| Rename column | `rename_{old}_to_{new}_on_{table}` |

## Checklist

- [ ] Up migration (PostgreSQL)
- [ ] Down migration — harus bisa rollback
- [ ] Entity updated (jika field berubah)
- [ ] Model updated (jika field berubah)
- [ ] Converter updated (jika field berubah)

## Referensi Rules

- [migration.md](../rules/migration.md)
- [entity.md](../rules/entity.md)
