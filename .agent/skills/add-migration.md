# Skill: Add Database Migration

## Trigger

Gunakan skill ini ketika perlu menambah atau mengubah tabel database (create table, add column, add index, dll).

## Prerequisites

- Deskripsi perubahan database yang diinginkan
- Driver yang di-support (MySQL, PostgreSQL, atau keduanya)

## Steps

### Step 1: Generate Timestamp

Gunakan format `YYYYMMDDHHmmss` berdasarkan waktu saat ini.

Contoh: `20260218100000`

### Step 2: Create MySQL Migration (jika diperlukan)

**File**: `db/migrations/{timestamp}_{description}.up.sql`

Contoh — Create table:

```sql
CREATE TABLE {table_name}
(
    id          VARCHAR(100) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NULL,
    parent_id   VARCHAR(100) NOT NULL,
    created_at  BIGINT       NOT NULL,
    updated_at  BIGINT       NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY fk_{table}_{fk_column} ({fk_column}) REFERENCES {parent_table} (id)
) ENGINE = InnoDB;
```

Contoh — Add column:

```sql
ALTER TABLE {table_name} ADD COLUMN {column_name} {type} {constraint};
```

Contoh — Add index:

```sql
CREATE INDEX idx_{table}_{column} ON {table_name} ({column});
```

**File**: `db/migrations/{timestamp}_{description}.down.sql`

Contoh — Drop table:

```sql
DROP TABLE IF EXISTS {table_name};
```

Contoh — Remove column:

```sql
ALTER TABLE {table_name} DROP COLUMN {column_name};
```

### Step 3: Create PostgreSQL Migration

**File**: `db/migrations/{timestamp}_{description}_pg.up.sql`

Sama dengan MySQL tapi tanpa `ENGINE = InnoDB` dan syntax PostgreSQL:

- Foreign key: `CONSTRAINT fk_name FOREIGN KEY (col) REFERENCES table (col)`
- Boolean: `BOOLEAN` bukan `TINYINT(1)`

**File**: `db/migrations/{timestamp}_{description}_pg.down.sql`

### Step 4: Update Entity (jika perlu)

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

- [ ] MySQL up migration
- [ ] MySQL down migration (harus bisa rollback)
- [ ] PostgreSQL up migration
- [ ] PostgreSQL down migration
- [ ] Entity updated (jika field berubah)
- [ ] Model updated (jika field berubah)
- [ ] Converter updated (jika field berubah)

## Referensi Rules

- [migration.md](../rules/migration.md)
- [entity.md](../rules/entity.md)
