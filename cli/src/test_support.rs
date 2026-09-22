//! Test-only scaffolding shared by the module suites.
//!
//! `std::env::set_var` is process-global while `cargo test` runs the suite in
//! threads, so every test that lets a verb reach the machine-wide registry
//! takes an [`IsolatedEnv`] first: it holds one process-wide mutex for the
//! whole test and points `XDG_CONFIG_HOME` and `HOME` at a scratch directory.
//! Without it a test would write the real `~/.config/zharness/repos`, and two
//! tests would race on the environment.

use std::ffi::OsString;
use std::fs;
use std::path::Path;
use std::sync::{Mutex, MutexGuard};

use tempfile::TempDir;

use crate::installer::{self, run_update, UpdateOptions};

/// Serialises every test that lets a verb reach the registry.
static ENV_LOCK: Mutex<()> = Mutex::new(());

/// Holds `ENV_LOCK` for the whole test and points `XDG_CONFIG_HOME` and `HOME`
/// at a scratch directory.
pub struct IsolatedEnv {
    _lock: MutexGuard<'static, ()>,
    /// Held for its Drop: the scratch directory must outlive the test.
    _dir: TempDir,
    xdg: Option<OsString>,
    home: Option<OsString>,
}

impl IsolatedEnv {
    pub fn new() -> Self {
        let lock = ENV_LOCK.lock().unwrap_or_else(|e| e.into_inner());
        let dir = tempfile::tempdir().expect("scratch config dir");
        let xdg = std::env::var_os("XDG_CONFIG_HOME");
        let home = std::env::var_os("HOME");
        std::env::set_var("XDG_CONFIG_HOME", dir.path());
        std::env::set_var("HOME", dir.path());
        Self {
            _lock: lock,
            _dir: dir,
            xdg,
            home,
        }
    }
}

impl Drop for IsolatedEnv {
    fn drop(&mut self) {
        restore("XDG_CONFIG_HOME", self.xdg.take());
        restore("HOME", self.home.take());
    }
}

fn restore(key: &str, value: Option<OsString>) {
    match value {
        Some(v) => std::env::set_var(key, v),
        None => std::env::remove_var(key),
    }
}

/// A scratch repository root.
pub fn temp_repo() -> TempDir {
    tempfile::tempdir().expect("scratch repo")
}

pub fn write_file(root: &Path, rel: &str, content: &str) {
    let p = root.join(rel);
    if let Some(dir) = p.parent() {
        fs::create_dir_all(dir).unwrap_or_else(|e| panic!("mkdir {}: {e}", dir.display()));
    }
    fs::write(&p, content).unwrap_or_else(|e| panic!("write {rel}: {e}"));
}

pub fn read_file(root: &Path, rel: &str) -> String {
    fs::read_to_string(root.join(rel)).unwrap_or_else(|e| panic!("read {rel}: {e}"))
}

/// The embedded bytes of a doc-set path, as a `String`.
pub fn embedded_str(path: &str) -> String {
    String::from_utf8(
        crate::embedded::read_file(path)
            .unwrap_or_else(|| panic!("embed read {path}: file not embedded"))
            .to_vec(),
    )
    .expect("embedded doc is UTF-8")
}

pub fn install_ok(root: &Path) -> String {
    let mut out = String::new();
    installer::install(root, "test", &mut out).unwrap_or_else(|e| panic!("install: {e}\n{out}"));
    out
}

pub fn update_ok(root: &Path) -> String {
    let mut out = String::new();
    run_update(
        UpdateOptions {
            root: root.to_path_buf(),
            version: "test".to_string(),
            force: false,
        },
        &mut out,
    )
    .unwrap_or_else(|e| panic!("update: {e}\n{out}"));
    out
}

/// Every file under `root` with its sha256, sorted — the observable state a
/// verb must leave untouched when it refuses.
pub fn tree_snapshot(root: &Path) -> String {
    let mut rows = Vec::new();
    collect_files(root, root, &mut rows);
    rows.sort();
    rows.join("\n")
}

fn collect_files(root: &Path, dir: &Path, rows: &mut Vec<String>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for e in entries.flatten() {
        let p = e.path();
        if p.is_dir() {
            collect_files(root, &p, rows);
        } else if let Ok(b) = fs::read(&p) {
            let rel = p
                .strip_prefix(root)
                .unwrap_or(&p)
                .to_string_lossy()
                .into_owned();
            rows.push(format!("{rel}\u{0}{}", installer::sha(&b)));
        }
    }
}

/// Overwrite the registry with `roots`, creating its directory. The verbs
/// append through `register`; a test that needs a root no verb would write
/// (a deleted repository) sets the file directly.
pub fn set_registry(roots: &[String]) -> Result<(), String> {
    let p = installer::registry::registry_path()?;
    if let Some(dir) = p.parent() {
        fs::create_dir_all(dir).map_err(|e| format!("mkdir {}: {e}", dir.display()))?;
    }
    let body = if roots.is_empty() {
        String::new()
    } else {
        format!("{}\n", roots.join("\n"))
    };
    fs::write(&p, body).map_err(|e| format!("write {}: {e}", p.display()))
}
