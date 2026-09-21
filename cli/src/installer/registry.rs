//! The machine-wide registry of repositories zharness manages, one absolute
//! root per line. It is the only way `update --check --all` can find consumer
//! repositories, which carry no hook of their own.

use std::fs;
use std::path::{Path, PathBuf};

use super::{dedupe, write_file_atomic};

/// `$XDG_CONFIG_HOME/zharness/repos`, falling back to `~/.config`.
pub fn registry_path() -> Result<PathBuf, String> {
    let dir = match std::env::var("XDG_CONFIG_HOME") {
        Ok(d) if !d.is_empty() => PathBuf::from(d),
        _ => {
            let home = std::env::var("HOME")
                .map_err(|_| "cannot determine the home directory".to_string())?;
            PathBuf::from(home).join(".config")
        }
    };
    Ok(dir.join("zharness").join("repos"))
}

/// Resolve symlinks so one repository reached through two paths (macOS
/// `/var` vs `/private/var`, a symlinked checkout) is one entry.
pub fn canonical_root(root: &Path) -> PathBuf {
    fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf())
}

/// The recorded repository roots in file order.
pub fn registered() -> Result<Vec<String>, String> {
    let p = registry_path()?;
    let body = match fs::read_to_string(&p) {
        Ok(s) => s,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(e) => return Err(format!("read {}: {e}", p.display())),
    };
    let roots: Vec<String> = body
        .split('\n')
        .map(|l| l.trim())
        .filter(|l| !l.is_empty())
        .map(|l| l.to_string())
        .collect();
    Ok(dedupe(roots))
}

fn write_registry(roots: &[String]) -> Result<(), String> {
    let p = registry_path()?;
    let body = if roots.is_empty() {
        String::new()
    } else {
        format!("{}\n", roots.join("\n"))
    };
    write_file_atomic(&p, body.as_bytes())
}

/// Advisory: a read-only or missing home must never fail the verb that called
/// it, so errors become a warning line.
pub fn register(root: &Path, out: &mut String) {
    let root = canonical_root(root).to_string_lossy().into_owned();
    let err = match registered() {
        Ok(roots) => {
            if roots.contains(&root) {
                return;
            }
            let mut next = roots;
            next.push(root);
            write_registry(&next).err()
        }
        Err(e) => Some(e),
    };
    if let Some(e) = err {
        out.push_str(&format!("warning    repo registry not updated: {e}\n"));
    }
}

pub fn unregister(root: &Path, out: &mut String) {
    let root = canonical_root(root).to_string_lossy().into_owned();
    let err = match registered() {
        Ok(roots) => {
            let kept: Vec<String> = roots.iter().filter(|r| **r != root).cloned().collect();
            if kept.len() == roots.len() {
                return;
            }
            write_registry(&kept).err()
        }
        Err(e) => Some(e),
    };
    if let Some(e) = err {
        out.push_str(&format!("warning    repo registry not updated: {e}\n"));
    }
}
