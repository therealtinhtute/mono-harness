//! The `docs/` → `.harness/` layout migration that `update` performs.

use super::check::check;
use super::ownership::{Ownership, KIND_FILE, ORIGIN_CREATED, ORIGIN_PREEXISTING};
use super::{load_base, run_update, sha, UpdateOptions, PROJECT_TARGET, WORKFLOW_TARGET};
use crate::test_support::{
    embedded_str, read_file, temp_repo, tree_snapshot, update_ok, write_file, IsolatedEnv,
};
use std::path::Path;

const ACTIVE: &str = "---\nstatus: active\n---\n# P\n\n## Goal\n- x\n\n## Phases and Verification\n- p\n\n## Log\n\n## Validation\n\n## Current State and Next Action\n- n\n";
const COMPLETED: &str = "# Done plan\r\nkept byte for byte \u{00e9}\n";
const PROJECT: &str = "# Project\n\nhand-written project facts\n";

/// A repository as a v0.30 binary left it: managed files under `docs/`, and
/// the base manifest and ownership ledger keyed on those paths.
fn old_layout(root: &Path) {
    write_file(root, "docs/WORKFLOW.md", "old router\n");
    write_file(root, "docs/PROJECT.md", PROJECT);
    write_file(root, "docs/plans/active/p.md", ACTIVE);
    write_file(root, "docs/plans/completed/q.md", COMPLETED);
    write_file(root, "AGENTS.md", "# Agents\n");
    write_file(
        root,
        ".zharness/base/manifest.json",
        &format!(
            "{{\"zharness_version\":\"0.30.0\",\"installed_at\":\"2026-10-01T00:00:00Z\",\"files\":[{{\"path\":\"docs/WORKFLOW.md\",\"sha256\":\"{}\"}}]}}\n",
            sha(b"old router\n")
        ),
    );
    write_file(
        root,
        ".zharness/base/ownership.tsv",
        "file\tdocs/WORKFLOW.md\tcreated_by_installer\nfile\tdocs/PROJECT.md\tpreexisting\n",
    );
}

fn without_bookkeeping(root: &Path) -> String {
    tree_snapshot(root)
        .lines()
        .filter(|l| !l.starts_with(".zharness/"))
        .collect::<Vec<_>>()
        .join("\n")
}

#[test]
fn update_moves_old_layout_into_harness_keeping_plan_bytes() {
    let _env = IsolatedEnv::new();
    let tmp = temp_repo();
    let root = tmp.path();
    old_layout(root);
    assert!(
        check(root)
            .unwrap()
            .iter()
            .any(|l| l.contains("docs/WORKFLOW.md")),
        "check must report the old layout as drift"
    );

    update_ok(root);

    assert_eq!(WORKFLOW_TARGET, ".harness/WORKFLOW.md");
    assert_eq!(PROJECT_TARGET, ".harness/PROJECT.md");
    assert_eq!(
        read_file(root, WORKFLOW_TARGET),
        embedded_str("WORKFLOW.md")
    );
    assert_eq!(read_file(root, PROJECT_TARGET), PROJECT);
    assert_eq!(read_file(root, ".harness/plans/active/p.md"), ACTIVE);
    assert_eq!(read_file(root, ".harness/plans/completed/q.md"), COMPLETED);
    for gone in ["docs/WORKFLOW.md", "docs/PROJECT.md", "docs/plans"] {
        assert!(!root.join(gone).exists(), "{gone} left behind");
    }
    assert!(read_file(root, "AGENTS.md").contains("<!-- ZHARNESS:BEGIN -->"));

    let base = load_base(root).unwrap();
    assert!(base.contains_key(WORKFLOW_TARGET));
    assert!(base.keys().all(|k| !k.starts_with("docs/")), "{base:?}");
    let own = Ownership::load(root);
    assert_eq!(own.get(KIND_FILE, WORKFLOW_TARGET), ORIGIN_CREATED);
    assert_eq!(own.get(KIND_FILE, PROJECT_TARGET), ORIGIN_PREEXISTING);
    assert!(!own.has(KIND_FILE, "docs/WORKFLOW.md"));
    // The fixture PROJECT.md is hand-written and lacks the template headings;
    // that drift is the project's to fix and says nothing about the move.
    let drift: Vec<String> = check(root)
        .unwrap()
        .into_iter()
        .filter(|l| !l.starts_with("heading "))
        .collect();
    assert_eq!(drift, Vec::<String>::new());

    let before = without_bookkeeping(root);
    let second = update_ok(root);
    assert!(!second.contains("moved"), "{second}");
    assert_eq!(without_bookkeeping(root), before);
}

#[test]
fn update_refuses_when_old_and_new_paths_both_exist() {
    let _env = IsolatedEnv::new();
    let tmp = temp_repo();
    let root = tmp.path();
    old_layout(root);
    write_file(root, ".harness/PROJECT.md", "a second copy\n");
    let before = tree_snapshot(root);

    let mut out = String::new();
    let err = run_update(
        UpdateOptions {
            root: root.to_path_buf(),
            version: "test".to_string(),
            force: false,
        },
        &mut out,
    )
    .expect_err("both copies present must refuse");

    assert!(
        err.contains("docs/PROJECT.md") && err.contains(".harness/PROJECT.md"),
        "{err}"
    );
    assert_eq!(tree_snapshot(root), before, "a refused update wrote");
}

#[test]
fn update_moves_then_migrates_a_legacy_plan_in_its_new_place() {
    let _env = IsolatedEnv::new();
    let tmp = temp_repo();
    let root = tmp.path();
    old_layout(root);
    let legacy: String = [
        "Outcome",
        "Authority and Requirements",
        "Non-goals",
        "Approach and Risks",
        "Phases and Verification",
        "Progress",
        "Decisions",
        "Validation",
        "Current State and Next Action",
    ]
    .iter()
    .map(|h| format!("## {h}\n- x\n\n"))
    .collect();
    write_file(root, "docs/plans/active/p.md", &format!("# P\n\n{legacy}"));

    let out = update_ok(root);

    assert!(
        out.contains("migrated       .harness/plans/active/p.md"),
        "{out}"
    );
    assert!(read_file(root, ".harness/plans/active/p.md").contains("## Log\n"));
    assert!(!root.join("docs/plans").exists());
}
