# PhotoTransfer LAN Architecture

## System Overview

```mermaid
flowchart LR
  Phone["iPhone / Android Browser"] -->|HTTPS optional / HTTP LAN| API["Go Fiber REST API"]
  Phone -->|WebSocket progress| WS["WebSocket Hub"]
  Desktop["Tauri Windows Desktop"] --> UI["Next.js App Router UI"]
  UI -->|REST| API
  UI -->|WebSocket| WS
  API --> App["Application Services"]
  App --> Domain["Domain Entities + Repositories"]
  App --> Store["Streaming Storage"]
  App --> Hash["SHA256 Verification"]
  Domain --> DB["SQLite via GORM"]
  Store --> Files["Organized Upload Folder"]
```

## Clean Architecture Boundaries

- `domain`: core entities, repository interfaces, upload statuses, domain errors.
- `application`: use cases such as first setup, authentication, upload session creation, chunk append, verification, duplicate handling.
- `infrastructure`: SQLite/GORM repositories, file storage, hashing, settings persistence.
- `presentation`: Fiber REST handlers, middleware, Swagger, WebSocket hub.
- `frontend`: Next.js desktop dashboard and mobile upload UI.
- `src-tauri`: Windows desktop shell, process bootstrap, installer config.

## Upload Sequence

```mermaid
sequenceDiagram
  participant M as Mobile Browser
  participant A as REST API
  participant S as Storage
  participant D as SQLite
  participant W as WebSocket

  M->>A: POST /api/upload-sessions
  A->>D: Create pending upload row
  A-->>M: session_id, chunk_size
  loop each 5MB chunk
    M->>A: PUT /api/upload-sessions/{id}/chunks/{index}
    A->>S: Append chunk stream
    A->>W: Publish progress
  end
  M->>A: POST /api/upload-sessions/{id}/complete
  A->>S: Compute SHA256
  A->>D: Detect duplicate and persist result
  A-->>M: Success / Failed / Corrupted
```

## Security Model

- Passwords are stored with bcrypt hashes only.
- Browser sessions use secure, httpOnly cookies when HTTPS is enabled.
- CSRF token is required for mutating browser requests.
- Rate limiting protects login and upload-session endpoints.
- Temporary QR token can be generated for short-lived mobile upload access.
- Auto logout is controlled by `settings.session_timeout_minutes`.

