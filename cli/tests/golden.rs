//! Golden replay: every fixture under `cli/testdata/golden/` is replayed
//! against the Rust binary and diffed byte-for-byte.
//!
//! The scenario setups below mirror `cli/testdata/golden/capture.sh` exactly —
//! same scratch layout, same isolated HOME/XDG_CONFIG_HOME, same seeded bytes.
//! The normalization is the same too: the scenario root becomes `<SCRATCH>`
//! and every `installed_at` value becomes `<TIMESTAMP>`.
//!
//! A divergence in success stdout, exit code, or `tree.sha256` is a bug. Only
//! help/usage/error-message text may be listed in a scenario's `allowed_break`
//! — and `exit` and `tree.sha256` are never allowed to appear there.

use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

use sha2::{Digest, Sha256};

const BROWN_FIELD_AGENTS: &str = "# Agents\n\nConsumer-authored guidance that predates zharness.\n";

const BROWN_FIELD_PROJECT: &str = "\
# PROJECT — identity

## What is this project?
- A consumer repository that predates zharness.

## Who is it for?
- The team that owns it.

## Non-goals
- Everything zharness does not manage.

## What are the gate commands?
- run from: repository root
- tests: `make test`
- types: n/a
- lint: `make lint`
- build: `make build`
- format: n/a

## Architecture in one breath
- runtime shape: one service
- where state lives: git
- what are the entrypoints: `cmd/serve`

## What are we working on right now?
- plan: docs/plans/active/consumer.md (active)
";

/// One replayed invocation.
struct Scenario {
    name: &'static str,
    /// The exit code the fixture pins.
    expected_exit: i32,
    /// Working directory relative to the scenario root.
    cwd: &'static str,
    args: &'static [&'static str],
    /// Builds the scenario root: `home/`, `xdg/` and the repositories.
    setup: fn(&Path),
    /// Fixture files whose content may differ. Only ever help/usage/error
    /// text; `exit` and `tree.sha256` are rejected at runtime.
    allowed_break: &'static [&'static str],
}

fn scenarios() -> Vec<Scenario> {
    vec![
        Scenario {
            name: "install-greenfield",
            expected_exit: 0,
            cwd: "repo",
            args: &["install"],
            setup: |root| new_repo(&root.join("repo")),
            allowed_break: &[],
        },
        Scenario {
            name: "install-brownfield",
            expected_exit: 0,
            cwd: "repo",
            args: &["install"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                brownfield(&repo);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "install-twice",
            expected_exit: 0,
            cwd: "repo",
            args: &["install"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-clean",
            expected_exit: 0,
            cwd: "repo",
            args: &["update"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-refuse-edited-block",
            expected_exit: 1,
            cwd: "repo",
            args: &["update"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                edit_agents_block(&repo);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-force",
            expected_exit: 0,
            cwd: "repo",
            args: &["update", "--force"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                edit_agents_block(&repo);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-check-clean",
            expected_exit: 0,
            cwd: "repo",
            args: &["update", "--check"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-check-drift",
            expected_exit: 1,
            cwd: "repo",
            args: &["update", "--check"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                append(&repo.join("docs/playbooks/work.md"), "\n<!-- drift -->\n");
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-check-all",
            expected_exit: 1,
            cwd: "repo-a",
            args: &["update", "--check", "--all"],
            setup: |root| {
                let a = root.join("repo-a");
                let b = root.join("repo-b");
                new_repo(&a);
                new_repo(&b);
                run_silent(root, &a, &["install"]);
                run_silent(root, &b, &["install"]);
                append(&b.join("docs/playbooks/work.md"), "\n<!-- drift -->\n");
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-check-all-clean",
            expected_exit: 0,
            cwd: "repo-a",
            args: &["update", "--check", "--all"],
            setup: |root| {
                let a = root.join("repo-a");
                let b = root.join("repo-b");
                new_repo(&a);
                new_repo(&b);
                run_silent(root, &a, &["install"]);
                run_silent(root, &b, &["install"]);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-all-without-check",
            expected_exit: 1,
            cwd: "repo",
            args: &["update", "--all"],
            setup: |root| new_repo(&root.join("repo")),
            allowed_break: &[],
        },
        Scenario {
            name: "update-check-force",
            expected_exit: 1,
            cwd: "repo",
            args: &["update", "--check", "--force"],
            setup: |root| new_repo(&root.join("repo")),
            allowed_break: &[],
        },
        Scenario {
            name: "uninstall-clean",
            expected_exit: 0,
            cwd: "repo",
            args: &["uninstall"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
            },
            allowed_break: &[],
        },
        Scenario {
            name: "uninstall-locally-modified",
            expected_exit: 0,
            cwd: "repo",
            args: &["uninstall"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                append(
                    &repo.join("docs/playbooks/work.md"),
                    "\n<!-- local edit -->\n",
                );
            },
            allowed_break: &[],
        },
        Scenario {
            name: "root-flag-outside-repo",
            expected_exit: 0,
            cwd: "outside",
            args: &["install", "--root", "../repo"],
            setup: |root| {
                new_repo(&root.join("repo"));
                fs::create_dir_all(root.join("outside")).expect("mkdir outside");
            },
            allowed_break: &[],
        },
        Scenario {
            name: "update-root-outside-repo",
            expected_exit: 0,
            cwd: "outside",
            args: &["update", "--root", "../repo"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                fs::create_dir_all(root.join("outside")).expect("mkdir outside");
            },
            allowed_break: &[],
        },
        Scenario {
            name: "uninstall-root-outside-repo",
            expected_exit: 0,
            cwd: "outside",
            args: &["uninstall", "--root", "../repo"],
            setup: |root| {
                let repo = root.join("repo");
                new_repo(&repo);
                run_silent(root, &repo, &["install"]);
                fs::create_dir_all(root.join("outside")).expect("mkdir outside");
            },
            allowed_break: &[],
        },
    ]
}

// ------------------------------------------------------------- helpers ---

fn fixture_root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("testdata/golden")
}

fn new_repo(p: &Path) {
    fs::create_dir_all(p).expect("mkdir repo");
    let out = Command::new("git")
        .args(["init", "-q", "-b", "main"])
        .current_dir(p)
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .output()
        .expect("run git init");
    assert!(out.status.success(), "git init failed in {}", p.display());
}

fn brownfield(repo: &Path) {
    fs::write(repo.join("AGENTS.md"), BROWN_FIELD_AGENTS).expect("write AGENTS.md");
    fs::create_dir_all(repo.join("docs")).expect("mkdir docs");
    fs::write(repo.join("docs/PROJECT.md"), BROWN_FIELD_PROJECT).expect("write PROJECT.md");
}

/// Insert a line inside the marked block, as a hand edit would.
fn edit_agents_block(repo: &Path) {
    let p = repo.join("AGENTS.md");
    let body = fs::read_to_string(&p).expect("read AGENTS.md");
    let marker = "<!-- ZHARNESS:BEGIN -->\n";
    let at = body.find(marker).expect("marked block present") + marker.len();
    let edited = format!("{}hand-edited line\n{}", &body[..at], &body[at..]);
    fs::write(&p, edited).expect("write AGENTS.md");
}

fn append(p: &Path, extra: &str) {
    let mut body = fs::read_to_string(p).expect("read file to append to");
    body.push_str(extra);
    fs::write(p, body).expect("write appended file");
}

fn clean_env(cmd: &mut Command, root: &Path) {
    cmd.env_clear()
        .env("PATH", std::env::var("PATH").unwrap_or_default())
        .env("HOME", root.join("home"))
        .env("XDG_CONFIG_HOME", root.join("xdg"))
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("LC_ALL", "C");
}

/// The version `capture.sh` pins the Go binary to. The replay must run a Rust
/// binary built with the same value, or the manifest's `zharness_version`
/// field — which the fixture pins by hash — cannot match.
const GOLDEN_VERSION: &str = "0.0.0-golden";

/// Build the binary under test with `ZHARNESS_VERSION` pinned, into its own
/// target directory so it never thrashes the outer `cargo test` build.
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

fn golden_binary() -> &'static Path {
    &GOLDEN_BIN
}

fn run_zharness(root: &Path, cwd: &Path, args: &[&str]) -> (String, String, i32) {
    let mut cmd = Command::new(golden_binary());
    cmd.args(args).current_dir(cwd);
    clean_env(&mut cmd, root);
    let out = cmd.output().expect("run zharness");
    (
        String::from_utf8_lossy(&out.stdout).into_owned(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
        out.status.code().unwrap_or(-1),
    )
}

fn run_silent(root: &Path, cwd: &Path, args: &[&str]) {
    let (_, _, code) = run_zharness(root, cwd, args);
    assert_eq!(code, 0, "setup `{}` failed", args.join(" "));
}

/// The scenario root becomes `<SCRATCH>` and every `installed_at` value
/// becomes `<TIMESTAMP>` — the same normalization `capture.sh` applies.
///
/// `zharness_version` is deliberately NOT tokenized: `capture.sh` pins it by
/// building Go with `-X main.version=0.0.0-golden`, and the replay pins it the
/// same way by building the Rust binary with `ZHARNESS_VERSION` set to the
/// same value (see [`golden_binary`]). Tokenizing it would let a port that
/// wrote the wrong version pass.
fn normalize(s: &str, root: &Path) -> String {
    let with_root = s.replace(&root.to_string_lossy().into_owned(), "<SCRATCH>");
    tokenize_json_string(&with_root, "\"installed_at\":\"", "<TIMESTAMP>")
}

/// Replace the value of a `"key":"value"` pair with `token`, leaving the key
/// and its quotes in place.
fn tokenize_json_string(s: &str, key: &str, token: &str) -> String {
    let mut out = String::with_capacity(s.len());
    let mut rest = s;
    while let Some(i) = rest.find(key) {
        let start = i + key.len();
        out.push_str(&rest[..start]);
        match rest[start..].find('"') {
            Some(end) => {
                out.push_str(token);
                rest = &rest[start + end..];
            }
            None => {
                rest = &rest[start..];
                break;
            }
        }
    }
    out.push_str(rest);
    out
}

fn sha256_hex(b: &[u8]) -> String {
    let mut h = Sha256::new();
    h.update(b);
    h.finalize().iter().map(|x| format!("{x:02x}")).collect()
}

/// Every path under `root` except `.git`, sorted, with normalized content
/// hashes — the same listing `capture.sh` writes.
fn tree(root: &Path, scen_root: &Path) -> String {
    let mut entries: Vec<(String, bool)> = Vec::new();
    walk(root, root, &mut entries);
    entries.sort();
    let mut out = String::new();
    for (rel, is_dir) in entries {
        if is_dir {
            out.push_str(&format!("dir  {rel}/\n"));
        } else {
            let bytes = fs::read(root.join(&rel)).expect("read fixture file");
            let norm = normalize(&String::from_utf8_lossy(&bytes), scen_root);
            out.push_str(&format!("{}  {rel}\n", sha256_hex(norm.as_bytes())));
        }
    }
    out
}

fn walk(base: &Path, dir: &Path, out: &mut Vec<(String, bool)>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for e in entries.flatten() {
        let name = e.file_name().to_string_lossy().into_owned();
        if name == ".git" {
            continue;
        }
        let p = e.path();
        let rel = p
            .strip_prefix(base)
            .expect("path under base")
            .to_string_lossy()
            .into_owned();
        if p.is_dir() {
            out.push((rel, true));
            walk(base, &p, out);
        } else {
            out.push((rel, false));
        }
    }
}

fn read_fixture(name: &str, file: &str) -> String {
    let p = fixture_root().join(name).join(file);
    fs::read_to_string(&p).unwrap_or_else(|e| panic!("read {}: {e}", p.display()))
}

/// The version the binary under test reports.
fn binary_version() -> String {
    let out = Command::new(golden_binary())
        .arg("--version")
        .output()
        .expect("run zharness --version");
    assert!(out.status.success(), "--version must exit 0");
    String::from_utf8_lossy(&out.stdout)
        .trim()
        .rsplit(' ')
        .next()
        .unwrap_or_default()
        .to_string()
}

/// Every manifest the run wrote must record the version the binary reports.
/// The replay tokenizes that field, so this is what keeps the plumbing pinned.
fn assert_manifest_versions(root: &Path, version: &str) -> usize {
    let mut found = 0usize;
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for e in entries.flatten() {
            let p = e.path();
            if p.is_dir() {
                if p.file_name().map(|n| n != ".git").unwrap_or(false) {
                    stack.push(p);
                }
            } else if p.file_name().map(|n| n == "manifest.json").unwrap_or(false) {
                let body = fs::read_to_string(&p).expect("read manifest");
                let want = format!("\"zharness_version\":\"{version}\"");
                assert!(
                    body.contains(&want),
                    "{}: expected {want} in {body}",
                    p.display()
                );
                found += 1;
            }
        }
    }
    found
}

// ---------------------------------------------------------------- tests ---

#[test]
fn golden_replay() {
    let mut failures: Vec<String> = Vec::new();
    let mut replayed = 0usize;
    let mut manifests = 0usize;
    let version = binary_version();

    for s in scenarios() {
        for f in s.allowed_break {
            assert!(
                *f != "exit" && *f != "tree.sha256",
                "{}: allowed_break may not cover {f} — exit codes and written bytes are parity",
                s.name
            );
        }

        let tmp = tempfile::tempdir().expect("temp dir");
        let root = tmp.path().canonicalize().expect("canonicalize scratch");
        fs::create_dir_all(root.join("home")).expect("mkdir home");
        fs::create_dir_all(root.join("xdg")).expect("mkdir xdg");
        (s.setup)(&root);

        let cwd = root.join(s.cwd);
        let (stdout, stderr, code) = run_zharness(&root, &cwd, s.args);
        replayed += 1;
        manifests += assert_manifest_versions(&root, &version);

        let cmd_line = format!("{} $ zharness {}\n", s.cwd, s.args.join(" "));
        let actual = [
            ("cmd", cmd_line),
            ("stdout", normalize(&stdout, &root)),
            ("stderr", normalize(&stderr, &root)),
            ("exit", format!("{code}\n")),
            ("tree.sha256", tree(&root, &root)),
        ];
        for (file, got) in actual {
            let want = read_fixture(s.name, file);
            if got != want && !s.allowed_break.contains(&file) {
                failures.push(format!(
                    "{}: {file} differs\n--- expected ---\n{want}\n--- actual ---\n{got}",
                    s.name
                ));
            }
        }
        assert_eq!(
            code, s.expected_exit,
            "{}: expected exit {}, got {code}",
            s.name, s.expected_exit
        );
    }

    assert_eq!(replayed, 17, "every scenario must be replayed");
    assert!(
        manifests > 0,
        "no scenario wrote a manifest, so the version plumbing is unverified"
    );
    assert!(
        failures.is_empty(),
        "{} fixture divergence(s):\n\n{}",
        failures.len(),
        failures.join("\n\n")
    );
}
