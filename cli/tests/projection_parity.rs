//! Projection parity: the embedded `WORKFLOW.md` and `playbooks/*.md` must be
//! byte-identical to their projection under the repository's `docs/`, in both
//! directions.
//!
//! `AGENTS.md` is embedded too but is not part of this projection — its
//! counterpart is the repo-root `AGENTS.md`, authored separately — so it is
//! out of scope here.
//!
//! Reading a missing projection root is a hard failure, not a skip: a skip
//! lets the gate pass silently, which is how the invariant was lost once
//! already.

use std::collections::HashSet;
use std::fs;
use std::path::{Path, PathBuf};

use zharness::embedded;

/// The checked-out `docs/` tree the managed docs are projected into.
fn projection_root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../docs")
}

/// The embedded side of the projection: `WORKFLOW.md` plus every playbook.
fn embedded_projection() -> Vec<(String, Vec<u8>)> {
    let mut out = vec![(
        "WORKFLOW.md".to_string(),
        embedded::read_file("WORKFLOW.md")
            .expect("WORKFLOW.md is embedded")
            .to_vec(),
    )];
    for name in embedded::playbook_names() {
        let rel = format!("playbooks/{name}");
        out.push((
            rel.clone(),
            embedded::read_file(&rel)
                .expect("every playbook name resolves")
                .to_vec(),
        ));
    }
    out
}

/// Collects every mismatch into `errors` rather than failing the caller, so
/// the drift test can assert on them.
fn compare_projection(embedded_set: &[(String, Vec<u8>)], root: &Path, errors: &mut Vec<String>) {
    for (rel, want) in embedded_set {
        let p = root.join(rel);
        match fs::read(&p) {
            Err(e) => errors.push(format!("{rel}: missing projection at {}: {e}", p.display())),
            Ok(got) if got != *want => {
                errors.push(format!("{rel}: embedded and projected copies differ"))
            }
            Ok(_) => {}
        }
    }

    let mut seen: HashSet<&str> = HashSet::new();
    for (rel, _) in embedded_set {
        if let Some(name) = rel.strip_prefix("playbooks/") {
            seen.insert(name);
        }
    }
    match fs::read_dir(root.join("playbooks")) {
        Err(e) => errors.push(format!("read projected playbooks/: {e}")),
        Ok(entries) => {
            for e in entries.flatten() {
                if e.path().is_dir() {
                    continue;
                }
                let name = e.file_name().to_string_lossy().into_owned();
                if !seen.contains(name.as_str()) {
                    errors.push(format!(
                        "playbooks/{name}: projected file has no embedded counterpart"
                    ));
                }
            }
        }
    }
}

#[test]
fn projection_parity() {
    let root = projection_root();
    assert!(
        root.is_dir(),
        "projection root {} unreadable",
        root.display()
    );
    let mut errors = Vec::new();
    compare_projection(&embedded_projection(), &root, &mut errors);
    assert!(
        errors.is_empty(),
        "embedded and projected docs have drifted:\n{}",
        errors.join("\n")
    );
}

/// Proves the comparator above actually reports a mismatch instead of passing
/// vacuously on a green run with nothing to catch.
#[test]
fn projection_parity_detects_drift() {
    let embedded_set = vec![
        ("WORKFLOW.md".to_string(), b"same\n".to_vec()),
        ("playbooks/watzup.md".to_string(), b"drifted\n".to_vec()),
    ];
    let tmp = tempfile::tempdir().expect("temp dir");
    fs::write(tmp.path().join("WORKFLOW.md"), b"same\n").expect("write WORKFLOW.md");
    fs::create_dir_all(tmp.path().join("playbooks")).expect("mkdir playbooks");
    fs::write(tmp.path().join("playbooks/watzup.md"), b"original\n").expect("write playbook");

    let mut errors = Vec::new();
    compare_projection(&embedded_set, tmp.path(), &mut errors);

    assert!(
        !errors.is_empty(),
        "expected compare_projection to report the drifted file, got no failure"
    );
    assert!(
        errors.iter().any(|m| m.contains("playbooks/watzup.md")),
        "expected a failure naming playbooks/watzup.md, got: {errors:?}"
    );
}
