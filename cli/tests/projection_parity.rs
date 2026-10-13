//! The embedded WORKFLOW router must match its repository projection.
//! AGENTS is authored separately. Procedures belong to installed skills.
//! Missing projections are hard failures, never a skip.
use std::fs;
use std::path::{Path, PathBuf};
use zharness::embedded;

fn projection_root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../docs")
}

fn embedded_projection() -> Vec<(String, Vec<u8>)> {
    vec![(
        "WORKFLOW.md".to_string(),
        embedded::read_file("WORKFLOW.md")
            .expect("WORKFLOW.md is embedded")
            .to_vec(),
    )]
}

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

#[test]
fn projection_parity_detects_drift() {
    let embedded_set = vec![("WORKFLOW.md".to_string(), b"upstream\n".to_vec())];
    let tmp = tempfile::tempdir().expect("temp dir");
    fs::write(tmp.path().join("WORKFLOW.md"), b"drifted\n").expect("write WORKFLOW.md");
    let mut errors = Vec::new();
    compare_projection(&embedded_set, tmp.path(), &mut errors);
    assert!(
        errors
            .iter()
            .any(|m| m == "WORKFLOW.md: embedded and projected copies differ"),
        "expected WORKFLOW drift, got {errors:?}"
    );
}

#[test]
fn projection_parity_detects_missing_projection() {
    let tmp = tempfile::tempdir().expect("temp dir");
    let mut errors = Vec::new();
    compare_projection(&embedded_projection(), tmp.path(), &mut errors);
    assert!(
        errors
            .iter()
            .any(|m| m.starts_with("WORKFLOW.md: missing projection")),
        "expected missing WORKFLOW failure, got {errors:?}"
    );
}
