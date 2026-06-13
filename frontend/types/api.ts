export type Dashboard = {
  local_ip: string;
  service_status: string;
  storage_location: string;
  available_space_bytes: number;
  upload_url: string;
  today_files: number;
  today_bytes: number;
};

export type LoginSession = {
  session_id: string;
  csrf_token: string;
  expires_at: string;
};

export type TemporaryToken = {
  token: string;
  upload_url: string;
  expires_at: string;
};

export type Settings = {
  id: number;
  upload_directory: string;
  auto_organize: boolean;
  max_upload_size: number | null;
  max_concurrent_uploads: number;
  session_timeout_minutes: number;
  auto_start_service: boolean;
};

export type ActivityLog = {
  id: number;
  event_type: string;
  actor: string;
  message: string;
  metadata_json: string;
  created_at: string;
};

export type UploadSession = {
  id: string;
  chunk_size: number;
};

export type UploadStatus = "pending" | "uploading" | "success" | "failed" | "corrupted" | "duplicate";

export type UploadRecord = {
  id: string;
  filename: string;
  original_filename: string;
  filesize: number;
  sha256: string;
  upload_time: string;
  device_name: string;
  status: UploadStatus;
  received_bytes: number;
};

export type SocketEvent =
  | { type: "upload_progress"; data: UploadRecord }
  | { type: "upload_complete"; data: UploadRecord };
