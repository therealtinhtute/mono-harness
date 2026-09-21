//! The command surface: clap definitions mirroring `internal/interfaces`, the
//! root resolution both `manage.go` and the verbs share, and the dispatch that
//! reproduces cobra's stdout/stderr split.
//!
//! Help, usage and error text differ from cobra's — that is a listed break
//! (plan R6), not a parity bug. Success stdout, exit codes and written bytes
//! are the parity contract.

use std::path::{Component, Path, PathBuf};

use clap::{Args, Parser, Subcommand};

use crate::installer;

/// The version string: `ZHARNESS_VERSION` when the release workflow sets it at
/// build time, else the crate version.
pub const VERSION: &str = match option_env!("ZHARNESS_VERSION") {
    Some(v) => v,
    None => env!("CARGO_PKG_VERSION"),
};

#[derive(Parser, Debug)]
#[command(
    name = "zharness",
    about = "Installer/updater for the markdown-first workflow harness",
    disable_version_flag = true
)]
pub struct Cli {
    /// version for zharness
    #[arg(short = 'v', long = "version")]
    pub version: bool,
    #[command(subcommand)]
    pub command: Option<Command>,
}

impl Cli {
    pub fn parse_args() -> Self {
        <Self as Parser>::parse()
    }
}

/// Render the root help, as cobra does for a bare invocation.
pub fn render_help() -> String {
    use clap::CommandFactory;
    Cli::command().render_help().to_string()
}

#[derive(Subcommand, Debug)]
pub enum Command {
    /// Scaffold the managed doc set into this repository (idempotent; brownfield report is read-only)
    Install(InstallArgs),
    /// Refresh playbooks, WORKFLOW.md and the AGENTS.md block; write docs/PROJECT.md only when absent
    Update(UpdateArgs),
    /// Remove the managed doc set only — consumer-owned bytes are never touched
    Uninstall(UninstallArgs),
}

#[derive(Args, Debug)]
pub struct InstallArgs {
    /// target repository root (default: git toplevel of cwd)
    #[arg(long, value_name = "string")]
    pub root: Option<String>,
}

#[derive(Args, Debug)]
pub struct UpdateArgs {
    /// target repository root (default: git toplevel of cwd)
    #[arg(long, value_name = "string")]
    pub root: Option<String>,
    /// replace an AGENTS.md block edited since the last write (default: refuse and print the diff)
    #[arg(long)]
    pub force: bool,
    /// report drift from this binary without writing; exit 1 on drift
    #[arg(long)]
    pub check: bool,
    /// with --check: every repository in ~/.config/zharness/repos
    #[arg(long)]
    pub all: bool,
}

#[derive(Args, Debug)]
pub struct UninstallArgs {
    /// target repository root (default: git toplevel of cwd)
    #[arg(long, value_name = "string")]
    pub root: Option<String>,
}

/// Run the requested verb. `Err` carries the message printed to stderr before
/// exiting 1.
pub fn dispatch(command: Command) -> Result<(), String> {
    match command {
        Command::Install(a) => {
            let root = resolve_root(a.root.as_deref());
            let mut out = String::new();
            match installer::install(&root, VERSION, &mut out) {
                Ok(()) => {
                    print!("{out}");
                    Ok(())
                }
                Err(err) => {
                    // install accumulates its report before failing, and the
                    // report belongs with the error rather than on stdout.
                    eprint!("{out}");
                    Err(err)
                }
            }
        }
        Command::Update(a) => {
            // Both refusals are decided before the root is resolved or any
            // file is read, exactly as manage.go orders them.
            if a.all && !a.check {
                return Err("--all requires --check".into());
            }
            if a.check && a.force {
                return Err("--check is read-only and takes no --force".into());
            }
            let root = resolve_root(a.root.as_deref());
            let mut out = String::new();
            let result = if a.check {
                installer::run_check(&root, a.all, &mut out)
            } else {
                installer::run_update(
                    installer::UpdateOptions {
                        root,
                        version: VERSION.to_string(),
                        force: a.force,
                    },
                    &mut out,
                )
            };
            print!("{out}");
            result
        }
        Command::Uninstall(a) => {
            let root = resolve_root(a.root.as_deref());
            let mut out = String::new();
            match installer::uninstall(&root, &mut out) {
                Ok(()) => {
                    print!("{out}");
                    Ok(())
                }
                Err(err) => {
                    eprint!("{out}");
                    Err(err)
                }
            }
        }
    }
}

/// Resolve the repository root for the managed-set verbs: an explicit `--root`
/// wins, else the git toplevel of the working directory, else the working
/// directory itself.
pub fn resolve_root(explicit: Option<&str>) -> PathBuf {
    if let Some(r) = explicit {
        if !r.is_empty() {
            return abs(Path::new(r));
        }
    }
    if let Ok(out) = std::process::Command::new("git")
        .args(["rev-parse", "--show-toplevel"])
        .output()
    {
        if out.status.success() {
            let top = String::from_utf8_lossy(&out.stdout).trim().to_string();
            if !top.is_empty() {
                return PathBuf::from(top);
            }
        }
    }
    abs(Path::new("."))
}

/// `filepath.Abs`: join against the working directory when relative, then
/// clean lexically. Deliberately does not resolve symlinks — `canonical_root`
/// is the function that does, and only for the registry.
pub fn abs(p: &Path) -> PathBuf {
    let joined = if p.is_absolute() {
        p.to_path_buf()
    } else {
        match std::env::current_dir() {
            Ok(cwd) => cwd.join(p),
            Err(_) => p.to_path_buf(),
        }
    };
    clean(&joined)
}

/// `filepath.Clean`: drop `.` components and resolve `..` lexically.
fn clean(p: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for c in p.components() {
        match c {
            Component::CurDir => {}
            Component::ParentDir => {
                if out.file_name().is_some() {
                    out.pop();
                } else if !out.has_root() {
                    out.push("..");
                }
            }
            other => out.push(other.as_os_str()),
        }
    }
    if out.as_os_str().is_empty() {
        out.push(".");
    }
    out
}
