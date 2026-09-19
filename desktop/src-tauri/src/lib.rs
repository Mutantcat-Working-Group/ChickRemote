mod config;
mod process;

use config::Settings;
use process::Engine;
use serde::Serialize;
use std::{
    fs::{self, OpenOptions},
    io::{Read, Seek, SeekFrom, Write},
    path::{Path, PathBuf},
    process::{Command, Stdio},
    sync::Mutex,
};
use tauri::{Manager, State};

#[derive(Default)]
struct Runtime {
    engine: Engine,
    port: Option<u16>,
}
type Shared = Mutex<Runtime>;

fn data_dir(app: &tauri::AppHandle) -> Result<PathBuf, String> {
    let dir = app.path().app_data_dir().map_err(|e| e.to_string())?;
    fs::create_dir_all(&dir).map_err(|e| e.to_string())?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&dir, fs::Permissions::from_mode(0o700)).map_err(|e| e.to_string())?;
    }
    Ok(dir)
}

fn private_write(path: &Path, data: &[u8]) -> Result<(), String> {
    let mut options = OpenOptions::new();
    options.write(true).create(true).truncate(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600);
    }
    let mut file = options.open(path).map_err(|e| e.to_string())?;
    file.write_all(data).map_err(|e| e.to_string())?;
    file.sync_all().map_err(|e| e.to_string())
}

#[tauri::command]
fn load_settings(app: tauri::AppHandle) -> Result<Settings, String> {
    let path = data_dir(&app)?.join("settings.json");
    match fs::read(path) {
        Ok(data) => serde_json::from_slice(&data).map_err(|e| e.to_string()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(Settings::default()),
        Err(e) => Err(e.to_string()),
    }
}

#[tauri::command]
fn save_settings(
    app: tauri::AppHandle,
    state: State<Shared>,
    settings: Settings,
) -> Result<(), String> {
    let mut runtime = state.lock().map_err(|e| e.to_string())?;
    if runtime.engine.running()? {
        return Err("Stop the client before changing settings".into());
    }
    settings.validate()?;
    private_write(
        &data_dir(&app)?.join("settings.json"),
        &serde_json::to_vec_pretty(&settings).map_err(|e| e.to_string())?,
    )
}

#[tauri::command]
fn start_client(app: tauri::AppHandle, state: State<Shared>) -> Result<(), String> {
    let mut runtime = state.lock().map_err(|e| e.to_string())?;
    if runtime.engine.running()? {
        return Err("Client is already running".into());
    }
    let settings = load_settings(app.clone())?;
    settings.validate()?;
    let listener = std::net::TcpListener::bind(("127.0.0.1", settings.dashboard_port))
        .map_err(|e| format!("Dashboard port unavailable: {e}"))?;
    let dir = data_dir(&app)?;
    let config_path = dir.join("client.yaml");
    // JSON is valid YAML, and structured serialization prevents YAML injection.
    private_write(
        &config_path,
        &serde_json::to_vec_pretty(&settings.engine_config(&dir)).map_err(|e| e.to_string())?,
    )?;
    let exe = std::env::current_exe().map_err(|e| e.to_string())?;
    let sidecar = exe
        .parent()
        .ok_or("Missing executable directory")?
        .join(if cfg!(windows) {
            "chickreomte-cli.exe"
        } else {
            "chickreomte-cli"
        });
    let log_path = dir.join("engine.log");
    private_write(&log_path, b"")?;
    let log = OpenOptions::new()
        .append(true)
        .open(&log_path)
        .map_err(|e| e.to_string())?;
    let mut command = Command::new(sidecar);
    command
        .arg("--conf")
        .arg(config_path)
        .env("CHICKREOMTE_DESKTOP", "1")
        .current_dir(&dir)
        .stdin(Stdio::null())
        .stderr(log.try_clone().map_err(|e| e.to_string())?)
        .stdout(log);
    if let Ok(appdir) = std::env::var("APPDIR") {
        command.env(
            "PATH",
            format!(
                "{appdir}/usr/bin:{}",
                std::env::var("PATH").unwrap_or_default()
            ),
        );
    }
    drop(listener);
    runtime.engine.start(command)?;
    runtime.port = Some(settings.dashboard_port);
    Ok(())
}

#[tauri::command]
fn stop_client(state: State<Shared>) -> Result<(), String> {
    let mut runtime = state.lock().map_err(|e| e.to_string())?;
    runtime.engine.stop()?;
    runtime.port = None;
    Ok(())
}

#[derive(Serialize)]
struct Status {
    running: bool,
    ready: bool,
    log: String,
    version: &'static str,
}

#[tauri::command]
fn client_status(app: tauri::AppHandle, state: State<Shared>) -> Result<Status, String> {
    let mut runtime = state.lock().map_err(|e| e.to_string())?;
    let running = runtime.engine.running()?;
    let ready = running
        && runtime
            .port
            .map(|port| {
                std::net::TcpStream::connect_timeout(
                    &std::net::SocketAddr::from(([127, 0, 0, 1], port)),
                    std::time::Duration::from_millis(50),
                )
                .is_ok()
            })
            .unwrap_or(false);
    let mut bytes = Vec::new();
    if let Ok(mut file) = fs::File::open(data_dir(&app)?.join("engine.log")) {
        let length = file.metadata().map_err(|e| e.to_string())?.len();
        file.seek(SeekFrom::Start(length.saturating_sub(16384)))
            .map_err(|e| e.to_string())?;
        file.take(16384)
            .read_to_end(&mut bytes)
            .map_err(|e| e.to_string())?;
    }
    Ok(Status {
        running,
        ready,
        log: String::from_utf8_lossy(&bytes).into_owned(),
        version: env!("CARGO_PKG_VERSION"),
    })
}

#[tauri::command]
fn open_dashboard(state: State<Shared>) -> Result<(), String> {
    let mut runtime = state.lock().map_err(|e| e.to_string())?;
    if !runtime.engine.running()? {
        return Err("Client is not running".into());
    }
    let port = runtime.port.ok_or("Dashboard is unavailable")?;
    open::that(format!("http://127.0.0.1:{port}")).map_err(|e| e.to_string())
}

pub fn run() {
    tauri::Builder::default()
        .manage(Shared::default())
        .invoke_handler(tauri::generate_handler![
            load_settings,
            save_settings,
            start_client,
            stop_client,
            client_status,
            open_dashboard
        ])
        .build(tauri::generate_context!())
        .expect("Failed to initialize ChickReomte")
        .run(|app, event| {
            if matches!(event, tauri::RunEvent::Exit) {
                if let Ok(mut runtime) = app.state::<Shared>().lock() {
                    let _ = runtime.engine.stop();
                }
            }
        });
}
