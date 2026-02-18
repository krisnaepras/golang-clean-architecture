# Agent Rules & Skills

Folder ini berisi aturan (rules) dan skill untuk AI coding agent agar tetap konsisten dengan arsitektur project ini, meskipun kode asli sudah di-refactor atau dihapus.

## Kompatibilitas

Rules dan skills ini kompatibel dengan:

- **GitHub Copilot** (membaca `.github/copilot-instructions.md` dan file AGENTS.md)
- **Antigravity / Claude Code** (membaca `AGENTS.md` dan `.agent/` folder)
- **Cursor** (membaca `.cursorrules` dan instruction files)
- **Windsurf / Codeium** (membaca `.windsurfrules`)
- **Aider** (membaca `.aider/conventions.md`)

## Struktur

```
.agent/
├── README.md                     # File ini
├── rules/
│   ├── architecture.md           # Aturan arsitektur clean architecture
│   ├── entity.md                 # Pattern untuk entity/domain layer
│   ├── model.md                  # Pattern untuk request/response model
│   ├── repository.md             # Pattern untuk repository layer
│   ├── usecase.md                # Pattern untuk usecase/business logic
│   ├── controller.md             # Pattern untuk HTTP controller
│   ├── messaging.md              # Pattern untuk Kafka producer/consumer
│   ├── config.md                 # Pattern untuk config & bootstrap
│   ├── migration.md              # Pattern untuk database migration
│   ├── testing.md                # Pattern untuk testing
│   └── converter.md              # Pattern untuk model converter
└── skills/
    ├── SKILL.md                  # Skill index
    ├── add-entity.md             # Skill: Menambah entity baru
    ├── add-crud-endpoint.md      # Skill: Menambah CRUD endpoint lengkap
    ├── add-kafka-event.md        # Skill: Menambah Kafka producer/consumer
    └── add-migration.md          # Skill: Menambah database migration
```
