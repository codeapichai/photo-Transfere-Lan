# PhotoTransfer LAN

Desktop Windows application for receiving photos and videos from iPhone or Android browsers over the same LAN without installing a mobile app.

## Download

Install the latest Windows build:

[Download PhotoTransfer LAN 0.1.5 for Windows](https://github.com/codeapichai/photo-Transfere-Lan/releases/download/v0.1.5/PhotoTransfer.LAN_0.1.5_x64_en-US.msi)

Installer files are published as GitHub Release assets instead of being stored in the source tree.

## Stack

- Next.js 15, TypeScript, TailwindCSS, App Router
- Tauri v2 desktop shell
- Go 1.25+, Fiber, SQLite, GORM
- REST API and WebSocket progress events

## Development

```powershell
cd backend
go run ./cmd/server
```

```powershell
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000` for the desktop UI and `http://localhost:3000/upload` for the mobile upload page.

## Build

```powershell
.\scripts\build-windows.ps1
```

The Windows installer is emitted under `src-tauri\target\release\bundle`.

## HTTPS Optional

By default the LAN service runs on HTTP port `8080`. To run with TLS, set both environment variables before starting the backend:

```powershell
$env:PT_HTTPS_CERT="C:\path\to\cert.pem"
$env:PT_HTTPS_KEY="C:\path\to\key.pem"
```

## Documents

- [Architecture](docs/architecture.md)
- [Folder Structure](docs/folder-structure.md)
- [Database Schema](docs/database-schema.sql)
- [OpenAPI Spec](docs/openapi.yaml)
- [Windows Clean Test Checklist](docs/windows-clean-test-checklist.md)
"# photo-Transfere-Lan" 
