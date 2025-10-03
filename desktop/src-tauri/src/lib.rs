use tauri_plugin_shell::ShellExt;

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[tauri::command]
fn check_default_messages_path() -> Option<String> {
    // Check if ~/Library/Messages/chat.db exists
    if let Some(home_dir) = dirs::home_dir() {
        let messages_path = home_dir.join("Library").join("Messages").join("chat.db");
        if messages_path.exists() {
            return messages_path.to_str().map(|s| s.to_string());
        }
    }
    None
}

#[tauri::command]
fn check_directory_exists(path: String) -> bool {
    std::path::Path::new(&path).is_dir()
}

#[tauri::command]
fn get_documents_dir() -> Option<String> {
    dirs::document_dir().and_then(|path| path.to_str().map(|s| s.to_string()))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .setup(|app| {
            // Start the Go backend server as a sidecar
            // Tauri automatically appends the target triple to the binary name
            let sidecar = app.shell().sidecar("threadbound")?;
            let (_rx, _child) = sidecar
                .args(["serve", "--port", "8765"])
                .spawn()?;

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![greet, check_default_messages_path, check_directory_exists, get_documents_dir])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
