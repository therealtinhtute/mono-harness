//! zharness — installer/updater for the markdown-first workflow harness.
//!
//! The three managed-set verbs (install / update / uninstall) own exactly the
//! managed doc set: `docs/WORKFLOW.md`, `docs/playbooks/*.md`, the marked
//! AGENTS block, the `docs/PROJECT.md` scaffold, `.zharness` bookkeeping and
//! the `.gitignore` entries. Nothing else in a repository is created,
//! modified, or deleted.

pub mod cli;
pub mod embedded;
pub mod installer;

#[cfg(test)]
mod test_support;

use std::process::ExitCode;

use clap::error::ErrorKind;

/// Parse the command line and run the requested verb, returning the process
/// exit code.
///
/// Exit codes are parity, not text: cobra prints help on stdout and exits 0
/// for a bare invocation, and exits 1 — not clap's default 2 — for a usage
/// error. Only the wording of help, usage and error output is a listed break
/// (plan R6).
pub fn run() -> ExitCode {
    let parsed = match <cli::Cli as clap::Parser>::try_parse() {
        Ok(c) => c,
        Err(e) => return clap_exit(e),
    };
    if parsed.version {
        println!("zharness version {}", cli::VERSION);
        return ExitCode::SUCCESS;
    }
    let Some(command) = parsed.command else {
        print!("{}", cli::render_help());
        return ExitCode::SUCCESS;
    };
    match cli::dispatch(command) {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("{err}");
            ExitCode::FAILURE
        }
    }
}

fn clap_exit(e: clap::Error) -> ExitCode {
    match e.kind() {
        ErrorKind::DisplayHelp | ErrorKind::DisplayVersion => {
            print!("{e}");
            ExitCode::SUCCESS
        }
        _ => {
            eprint!("{e}");
            ExitCode::FAILURE
        }
    }
}
