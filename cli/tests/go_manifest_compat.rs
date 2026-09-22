//! R4: the Rust binary must accept a `.zharness/base/manifest.json` written by
//! Go v0.23.1 without a forced reinstall.
//!
//! `cli/testdata/golden/go-installed/` is a brownfield repository installed by
//! the Go binary, captured raw: the manifest keeps the real RFC3339
//! `installed_at` Go wrote, and the ownership ledger carries both
//! `created_by_installer` and `preexisting` rows alongside the captured
//! originals. This test copies it, then drives `update`, `update --check` and
//! `uninstall` with the Rust binary and asserts the outcomes.

use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const GOLDEN_VERSION: &str = "0.0.0-golden";

static GOLDEN_BIN: std::sync::LazyLock<PathBuf> = std::sync::LazyLock::new(|| {
    let manifest_dir = Path::new(env!("CARGO_MANIFEST_DIR"));
    let target_dir = manifest_dir.join("target/golden-bin");
    let cargo = std::env::var("CARGO").unwrap_or_else(|_| "cargo".to_string());
    let out = Command::new(cargo)
        .args(["build", "--bin", "zharness", "--target-dir"])
        .arg(&target_dir)
        .env("ZHARNESS_VERSION", GOLDEN_VERSION)
        .current_dir(manifest_dir)
        .output()
        .expect("run cargo build for the golden binary");
    assert!(
        out.status.success(),
        "golden binary build failed:\n{}",
        String::from_utf8_lossy(&out.stderr)
    );
    target_dir.join("debug/zharness")
});

fn fixture() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("testdata/golden/go-installed")
}

/// A scratch copy of the Go-installed fixture, with an isolated HOME and
/// XDG_CONFIG_HOME beside it.
struct Scratch {
    _tmp: tempfile::TempDir,
    root: PathBuf,
    repo: PathBuf,
}

impl Scratch {
    fn new() -> Self {
        let tmp = tempfile::tempdir().expect("temp dir");
        let root = tmp.path().canonicalize().expect("canonicalize scratch");
        let repo = root.join("repo");
        copy_tree(&fixture(), &repo);
        fs::create_dir_all(root.join("home")).expect("mkdir home");
        fs::create_dir_all(root.join("xdg")).expect("mkdir xdg");
        Self {
            _tmp: tmp,
            root,
            repo,
        }
    }

    fn run(&self, args: &[&str]) -> (String, String, i32) {
        let mut cmd = Command::new(&*GOLDEN_BIN);
        cmd.args(args).current_dir(&self.repo);
        cmd.env_clear()
            .env("PATH", std::env::var("PATH").unwrap_or_default())
            .env("HOME", self.root.join("home"))
            .env("XDG_CONFIG_HOME", self.root.join("xdg"))
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("LC_ALL", "C");
        let out = cmd.output().expect("run zharness");
        (
            String::from_utf8_lossy(&out.stdout).into_owned(),
            String::from_utf8_lossy(&out.stderr).into_owned(),
            out.status.code().unwrap_or(-1),
        )
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.repo.join(rel)).unwrap_or_else(|e| panic!("read {rel}: {e}"))
    }

    fn exists(&self, rel: &str) -> bool {
        self.repo.join(rel).exists()
    }
}

fn copy_tree(src: &Path, dst: &Path) {
    fs::create_dir_all(dst).expect("mkdir dst");
    for e in fs::read_dir(src).expect("read src").flatten() {
        let from = e.path();
        let to = dst.join(e.file_name());
        if from.is_dir() {
            copy_tree(&from, &to);
        } else {
            fs::copy(&from, &to).expect("copy file");
        }
    }
}

/// The manifest's four schema fields, as Go wrote them.
fn manifest_fields(repo: &Path) -> (String, String, Vec<String>) {
    let body =
        fs::read_to_string(repo.join(".zharness/base/manifest.json")).expect("read manifest");
    let v: serde_json::Value = serde_json::from_str(&body).expect("manifest is valid JSON");
    let version = v["zharness_version"]
        .as_str()
        .expect("zharness_version is a string")
        .to_string();
    let installed_at = v["installed_at"]
        .as_str()
        .expect("installed_at is a string")
        .to_string();
    let paths: Vec<String> = v["files"]
        .as_array()
        .expect("files is an array")
        .iter()
        .map(|f| {
            assert!(f["sha256"].as_str().is_some(), "files[].sha256 is a string");
            f["path"]
                .as_str()
                .expect("files[].path is a string")
                .to_string()
        })
        .collect();
    (version, installed_at, paths)
}

#[test]
fn go_written_manifest_is_accepted() {
    let s = Scratch::new();

    // The fixture must be the raw Go artifact, not a normalized stand-in.
    let (version, installed_at, paths) = manifest_fields(&s.repo);
    assert_eq!(
        version, GOLDEN_VERSION,
        "fixture records the pinned version"
    );
    assert!(
        installed_at.len() == 20 && installed_at.ends_with('Z') && installed_at.contains('T'),
        "fixture installed_at must be a real RFC3339 stamp, got {installed_at:?}"
    );
    assert!(paths.contains(&"AGENTS.md".to_string()));
    assert!(paths.contains(&"docs/PROJECT.md".to_string()));

    // `update` must accept it without a forced reinstall.
    let (stdout, stderr, code) = s.run(&["update"]);
    assert_eq!(
        code, 0,
        "update refused the Go manifest:\n{stdout}\n{stderr}"
    );
    assert!(
        stdout.contains("update complete."),
        "update did not complete:\n{stdout}"
    );

    // The schema survives the rewrite: same fields, same shape.
    let (version2, installed_at2, paths2) = manifest_fields(&s.repo);
    assert_eq!(version2, GOLDEN_VERSION);
    assert!(
        installed_at2.len() == 20 && installed_at2.ends_with('Z'),
        "update wrote a non-RFC3339 installed_at: {installed_at2:?}"
    );
    assert_eq!(paths2, paths, "update changed the manifest's file set");

    // And the repository is now current against this binary.
    let (stdout, stderr, code) = s.run(&["update", "--check"]);
    assert_eq!(
        code, 0,
        "update --check reported drift:\n{stdout}\n{stderr}"
    );
    assert!(
        stdout.starts_with("current  "),
        "update --check did not report current:\n{stdout}"
    );
}

#[test]
fn go_installed_repo_uninstalls_cleanly() {
    let s = Scratch::new();

    // The consumer's bytes, as the Go binary captured them.
    let orig_agents = s.read(".zharness/base/original/AGENTS.md.orig");
    let orig_project = s.read(".zharness/base/original/docs_2FPROJECT.md.orig");

    let (stdout, stderr, code) = s.run(&["uninstall"]);
    assert_eq!(code, 0, "uninstall failed:\n{stdout}\n{stderr}");

    // The managed set is gone.
    assert!(
        !s.exists("docs/WORKFLOW.md"),
        "WORKFLOW.md survived uninstall"
    );
    assert!(!s.exists("docs/playbooks"), "playbooks/ survived uninstall");
    assert!(!s.exists(".zharness"), ".zharness/ survived uninstall");

    // The consumer's own bytes are back, byte-for-byte.
    assert_eq!(
        s.read("AGENTS.md"),
        orig_agents,
        "AGENTS.md was not restored to its pre-install original"
    );
    assert_eq!(
        s.read("docs/PROJECT.md"),
        orig_project,
        "docs/PROJECT.md was not restored to its pre-install original"
    );
}
