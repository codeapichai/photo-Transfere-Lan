PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS uploads (
  id TEXT PRIMARY KEY,
  filename TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  filesize INTEGER NOT NULL,
  sha256 TEXT,
  upload_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  device_name TEXT,
  status TEXT NOT NULL,
  storage_path TEXT NOT NULL,
  received_bytes INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_uploads_sha256 ON uploads(sha256);
CREATE INDEX IF NOT EXISTS idx_uploads_upload_time ON uploads(upload_time);

CREATE TABLE IF NOT EXISTS settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  upload_directory TEXT NOT NULL,
  auto_organize BOOLEAN NOT NULL DEFAULT 1,
  max_upload_size INTEGER,
  max_concurrent_uploads INTEGER NOT NULL DEFAULT 3,
  session_timeout_minutes INTEGER NOT NULL DEFAULT 30,
  auto_start_service BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS activity_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  actor TEXT,
  message TEXT NOT NULL,
  metadata_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  csrf_token TEXT NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS temporary_tokens (
  id TEXT PRIMARY KEY,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_temporary_tokens_expires_at ON temporary_tokens(expires_at);
