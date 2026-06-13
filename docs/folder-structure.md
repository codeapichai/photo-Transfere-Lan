# Folder Structure

```text
backend/
  cmd/server/                 Fiber server entrypoint
  internal/domain/            Entities and repository contracts
  internal/application/       Use cases and orchestration
  internal/infrastructure/    GORM, SQLite, filesystem, hashing
  internal/presentation/      HTTP handlers, middleware, WebSocket
  migrations/                 SQL schema migrations
frontend/
  app/                        Next.js App Router routes
  components/                 Reusable UI components
  lib/                        API and WebSocket clients
  types/                      Shared TypeScript types
src-tauri/
  src/                        Tauri Rust bootstrap
docs/
  architecture.md
  folder-structure.md
  database-schema.sql
  openapi.yaml
```

## Runtime Paths

- Default database: `%APPDATA%/PhotoTransferLAN/photo_transfer.db`
- Default uploads: `%USERPROFILE%/Pictures/PhotoTransferLAN`
- Logs: `%APPDATA%/PhotoTransferLAN/logs`

