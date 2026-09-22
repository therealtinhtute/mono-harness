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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::installer::uninstall;
    use crate::test_support::{install_ok, set_registry, temp_repo, update_ok, IsolatedEnv};

    /// `update` registers a root that is not already recorded.
    #[test]
    fn registry_update_registers() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        set_registry(&[]).unwrap();

        update_ok(root);

        let got = registered().unwrap();
        let want = canonical_root(root).to_string_lossy().into_owned();
        assert_eq!(got, vec![want], "registry = {got:?}");
    }

    /// One repository reached through two paths (a symlinked checkout) is one
    /// entry, and uninstalling through the link removes it.
    #[test]
    fn registry_symlinked_path_is_one_entry() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        let linkdir = temp_repo();
        let link = linkdir.path().join("link");
        std::os::unix::fs::symlink(root, &link).expect("symlink");

        install_ok(root);
        install_ok(&link);
        assert_eq!(
            registered().unwrap().len(),
            1,
            "registry = {:?}",
            registered().unwrap()
        );

        let mut out = String::new();
        uninstall(&link, &mut out).unwrap_or_else(|e| panic!("uninstall: {e}\n{out}"));
        assert!(
            registered().unwrap().is_empty(),
            "registry after uninstall via symlink = {:?}",
            registered().unwrap()
        );
    }

    /// A read-only or missing home must never fail the verb that called it:
    /// the registry warning is advisory and install still succeeds.
    #[test]
    fn registry_unwritable_home_warns_and_install_succeeds() {
        let _env = IsolatedEnv::new();
        let dir = temp_repo();
        // A regular file where the config dir should be makes create_dir_all
        // fail on every platform, including when the test runs as root.
        let blocker = dir.path().join("cfg");
        std::fs::write(&blocker, "x").unwrap();
        std::env::set_var("XDG_CONFIG_HOME", &blocker);

        let repo = temp_repo();
        let out = install_ok(repo.path());
        assert!(
            out.contains("warning    repo registry not updated"),
            "expected registry warning, got:\n{out}"
        );
    }
}
