# Agent Skills Index

Skills untuk membantu AI agent mengembangkan project Go Clean Architecture ini.

## Available Skills

| Skill             | File                                         | Deskripsi                                       |
| ----------------- | -------------------------------------------- | ----------------------------------------------- |
| Add Entity        | [add-entity.md](add-entity.md)               | Menambah domain entity baru beserta migration   |
| Add CRUD Endpoint | [add-crud-endpoint.md](add-crud-endpoint.md) | Menambah full CRUD endpoint (entity → route)    |
| Add Kafka Event   | [add-kafka-event.md](add-kafka-event.md)     | Menambah Kafka producer & consumer untuk entity |
| Add Migration     | [add-migration.md](add-migration.md)         | Menambah database migration file                |

## Cara Penggunaan

Refensi skill ini di prompt ke agent:

- "Gunakan skill add-entity untuk menambah entity Product"
- "Ikuti skill add-crud-endpoint untuk buat CRUD order"

Atau agent akan otomatis membaca skill ini ketika diminta menambah fitur baru.
