use once_cell::sync::OnceCell;
use std::{
    fs::{self, OpenOptions},
    process::{Child, Command, Stdio},
};
use tauri::{Manager, RunEvent};

static BACKEND_CHILD: OnceCell<std::sync::Mutex<Option<Child>>> = OnceCell::new();

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let child = start_backend_process()?;
            let _ = BACKEND_CHILD.set(std::sync::Mutex::new(Some(child)));

            if let Some(window) = app.get_webview_window("main") {
                let _ = window.set_focus();
            }

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building PhotoTransfer LAN")
        .run(|_app_handle, event| {
            if matches!(event, RunEvent::ExitRequested { .. } | RunEvent::Exit) {
                if let Some(child_slot) = BACKEND_CHILD.get() {
                    if let Ok(mut guard) = child_slot.lock() {
                        if let Some(child) = guard.take() {
                            let mut child = child;
                            let _ = child.kill();
                        }
                    }
                }
            }
        });
}

fn start_backend_process() -> Result<Child, Box<dyn std::error::Error>> {
    let exe_dir = std::env::current_exe()?
        .parent()
        .ok_or("cannot resolve application directory")?
        .to_path_buf();
    let backend = exe_dir.join("phototransfer-backend.exe");
    let log_dir = dirs::data_dir()
        .unwrap_or_else(std::env::temp_dir)
        .join("PhotoTransferLAN")
        .join("logs");
    fs::create_dir_all(&log_dir)?;
    let stdout = OpenOptions::new()
        .create(true)
        .append(true)
        .open(log_dir.join("backend.out.log"))?;
    let stderr = OpenOptions::new()
        .create(true)
        .append(true)
        .open(log_dir.join("backend.err.log"))?;

    Ok(Command::new(backend)
        .current_dir(exe_dir)
        .stdin(Stdio::null())
        .stdout(Stdio::from(stdout))
        .stderr(Stdio::from(stderr))
        .spawn()?)
}
