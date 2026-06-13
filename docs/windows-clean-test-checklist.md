# Windows Clean Machine Test Checklist

Use a Windows 10/11 machine or VM that has not been used for development.

## Install

1. Copy `PhotoTransfer LAN_0.1.0_x64_en-US.msi` to the machine.
2. Install with the default options.
3. Launch PhotoTransfer LAN from Start Menu.
4. Confirm Windows Defender/SmartScreen behavior. Unsigned builds may show a warning.

## First Launch

1. Open Setup.
2. Create a username and password.
3. Login.
4. Confirm the dashboard shows local IP, running service, storage folder, and QR code.

## Mobile Upload

1. Connect iPhone/Android and Windows PC to the same WiFi/LAN.
2. Scan the QR code.
3. Confirm the phone opens `http://<pc-ip>:8080/upload?token=...`.
4. Upload one JPG/PNG, one HEIC if available, and one MP4/MOV.
5. Confirm progress updates on phone and dashboard.
6. Confirm files appear in the configured upload folder by date.

## Security

1. Open dashboard in a new browser profile without logging in and confirm `/api/dashboard` returns `401`.
2. Try a mutating request without `X-CSRF-Token` and confirm `403`.
3. Wait for a temporary token to expire, then confirm upload session creation returns `401`.
4. Logout and confirm protected pages require login again.

## Settings And Logs

1. Change upload folder and upload a new file.
2. Confirm the new file is written to the new folder.
3. Toggle auto organize and confirm folder behavior.
4. Export logs as CSV.
5. Confirm login, token creation, upload, and settings events are present.

## Uninstall

1. Uninstall from Windows Settings.
2. Confirm application files are removed.
3. Confirm user data remains under `%APPDATA%/PhotoTransferLAN` unless manual cleanup is requested.
