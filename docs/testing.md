# Testing Plan

## Backend

- Unit tests for authentication, upload session validation, duplicate policy, and SHA256 verification.
- Integration tests using temporary SQLite databases and temporary upload folders.
- Upload tests stream file chunks through Fiber handlers and assert memory-safe append behavior.
- WebSocket tests subscribe to `/api/ws` and assert `upload_progress` and `upload_complete` events.

## Frontend

- Type checking with `npm run typecheck`.
- Component tests for setup, dashboard, and upload progress states.
- Browser tests for dashboard QR rendering and mobile upload flow.

## Coverage Target

The target is 80%+ for backend application services and presentation handlers. Infrastructure coverage should focus on repository and filesystem edge cases rather than generated boilerplate.

