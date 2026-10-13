//! M01–M11: real installer seams, bytes and existing recovery metadata.
use super::ownership::{Ownership, KIND_FILE, ORIGIN_CREATED, ORIGIN_PREEXISTING, OWNERSHIP_FILE};
use super::*;
use crate::test_support::{
    install_ok, read_file, temp_repo, tree_snapshot, update_ok, write_file, IsolatedEnv,
};

const WORK: &str = "docs/playbooks/work.md";
const FIRST: &str = "docs/playbooks/brainstorm.md";
const OLD: &str = "# previously shipped projection\n";
const ORIGINAL: &str = "# consumer original before installation\n";
const EDITED: &str = "# consumer edits after installation\n";

fn legacy(root: &Path, path: &str, origin: &str, original: Option<&str>) {
    write_file(root, path, OLD);
    let mut base = load_base(root).unwrap();
    base.insert(path.to_owned(), sha(OLD.as_bytes()));
    save_base(root, "previous", &base).unwrap();
    let mut own = Ownership::load(root);
    own.set(KIND_FILE, path, origin);
    own.save(root).unwrap();
    if let Some(bytes) = original {
        write_file(
            root,
            &format!("{ORIGINAL_DIR}/{}.orig", safe_path(path)),
            bytes,
        );
    }
}

fn assert_relinquished(root: &Path, path: &str) {
    assert!(
        !load_base(root).unwrap().contains_key(path),
        "retired active hash remains for {path}"
    );
    assert!(!all_targets().iter().any(|t| t.dst == path));
}

fn uninstall_ok(root: &Path) -> String {
    let mut out = String::new();
    uninstall(root, &mut out).unwrap_or_else(|e| panic!("uninstall: {e}\n{out}"));
    out
}

#[test]
fn m01_created_matching_retirement_repeat_and_uninstall() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_CREATED, None);
    let out = update_ok(root);
    assert!(
        !root.join(WORK).exists(),
        "matching created projection survived: {out}"
    );
    assert!(out.contains(WORK) && out.contains("removed"));
    assert_relinquished(root, WORK);
    assert!(!Ownership::load(root).has(KIND_FILE, WORK));
    update_ok(root);
    uninstall_ok(root);
    assert!(!root.join(WORK).exists());
    assert!(!Ownership::load(root).has(KIND_FILE, WORK));
}

#[test]
fn m02_edited_projection_and_original_survive_repeat_and_uninstall() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    write_file(root, WORK, EDITED);
    for _ in 0..2 {
        let out = update_ok(root);
        assert_eq!(read_file(root, WORK), EDITED, "{out}");
        assert!(out.contains(WORK) && (out.contains("modified") || out.contains("unknown")));
        assert_relinquished(root, WORK);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
    }
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), EDITED);
    assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
    assert_eq!(
        Ownership::load(root).get(KIND_FILE, WORK),
        ORIGIN_PREEXISTING
    );
}

#[test]
fn m03_missing_hash_or_unknown_origin_keeps_matching_bytes() {
    let _env = IsolatedEnv::new();
    for mode in ["missing-hash", "missing-base", "unknown"] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(
            root,
            WORK,
            if mode == "unknown" {
                "unknown"
            } else {
                ORIGIN_CREATED
            },
            None,
        );
        if mode == "missing-hash" {
            let mut base = load_base(root).unwrap();
            base.remove(WORK);
            save_base(root, "old", &base).unwrap();
        } else if mode == "missing-base" {
            fs::remove_file(root.join(MANIFEST_FILE)).unwrap();
        }
        for _ in 0..2 {
            let out = update_ok(root);
            assert_eq!(read_file(root, WORK), OLD, "{mode}: {out}");
            assert!(
                out.contains(WORK) && out.contains("unknown"),
                "{mode}: {out}"
            );
            assert_relinquished(root, WORK);
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
    }
}

#[test]
fn m04_corrupt_or_unreadable_manifest_fails_before_any_write() {
    let _env = IsolatedEnv::new();
    for directory in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_CREATED, None);
        fs::remove_file(root.join(MANIFEST_FILE)).unwrap();
        if directory {
            fs::create_dir(root.join(MANIFEST_FILE)).unwrap();
        } else {
            write_file(root, MANIFEST_FILE, "not json\n");
        }
        let before = tree_snapshot(root);
        let mut out = String::new();
        assert!(run_update(
            UpdateOptions {
                root: root.into(),
                version: "new".into(),
                force: false
            },
            &mut out
        )
        .is_err());
        assert_eq!(tree_snapshot(root), before);
        assert!(!out.contains("removed"));
    }
}

#[test]
fn m05_fresh_distribution_and_update_never_create_playbooks() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    assert!(!root.join("docs/playbooks").exists());
    write_file(root, PROJECT_TARGET, "consumer project\n");
    update_ok(root);
    assert!(!root.join("docs/playbooks").exists());
    assert_eq!(read_file(root, PROJECT_TARGET), "consumer project\n");
    assert!(read_file(root, AGENTS_TARGET).contains(BLOCK_BEGIN));
    uninstall_ok(root);
    assert_eq!(read_file(root, PROJECT_TARGET), "consumer project\n");
}

#[test]
fn m07_uninstall_old_shape_restores_original_and_preserves_edits() {
    let _env = IsolatedEnv::new();
    for edited in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        if edited {
            write_file(root, WORK, EDITED);
        }
        uninstall_ok(root);
        assert_eq!(
            read_file(root, WORK),
            if edited { EDITED } else { ORIGINAL }
        );
        if edited {
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        }
    }
}

#[test]
fn m08_unrelated_nested_content_and_ledger_survive() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_CREATED, None);
    write_file(root, "docs/playbooks/consumer/note.md", "unrelated\n");
    let mut own = Ownership::load(root);
    own.set(KIND_FILE, "unrelated/path", "preexisting");
    own.save(root).unwrap();
    for _ in 0..2 {
        update_ok(root);
        assert_eq!(
            read_file(root, "docs/playbooks/consumer/note.md"),
            "unrelated\n"
        );
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, "unrelated/path"),
            "preexisting"
        );
    }
    uninstall_ok(root);
    assert_eq!(
        read_file(root, "docs/playbooks/consumer/note.md"),
        "unrelated\n"
    );
    assert_eq!(
        Ownership::load(root).get(KIND_FILE, "unrelated/path"),
        "preexisting"
    );
}

#[test]
fn m09_preexisting_matching_restores_and_relinquishes_all_records() {
    let _env = IsolatedEnv::new();
    for legacy_name in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        if legacy_name {
            fs::rename(
                find_original(root, WORK).unwrap(),
                root.join(ORIGINAL_DIR)
                    .join(format!("{}.orig", legacy_safe_path(WORK))),
            )
            .unwrap();
        }
        let out = update_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL, "{out}");
        assert_relinquished(root, WORK);
        assert!(!Ownership::load(root).has(KIND_FILE, WORK));
        assert!(find_original(root, WORK).is_none());
        update_ok(root);
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert!(find_original(root, WORK).is_none());
    }
}

#[test]
fn m09_missing_or_contradictory_original_keeps_and_preserves_provenance() {
    let _env = IsolatedEnv::new();
    for contradictory in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let origin = if contradictory {
            ORIGIN_CREATED
        } else {
            ORIGIN_PREEXISTING
        };
        legacy(
            root,
            WORK,
            origin,
            if contradictory { Some(ORIGINAL) } else { None },
        );
        for _ in 0..2 {
            update_ok(root);
            assert_eq!(read_file(root, WORK), OLD);
            assert_relinquished(root, WORK);
            assert_eq!(Ownership::load(root).get(KIND_FILE, WORK), origin);
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
        assert_eq!(Ownership::load(root).get(KIND_FILE, WORK), origin);
        if contradictory {
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        }
    }
}

#[test]
fn m09_absent_ledger_seeds_but_malformed_existing_ledger_never_infers() {
    let _env = IsolatedEnv::new();
    for malformed in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        if malformed {
            write_file(
                root,
                OWNERSHIP_FILE,
                "malformed provenance retained verbatim\n",
            );
        } else {
            fs::remove_file(root.join(OWNERSHIP_FILE)).unwrap();
        }
        update_ok(root);
        assert_eq!(
            read_file(root, WORK),
            if malformed { OLD } else { ORIGINAL }
        );
        assert_relinquished(root, WORK);
        update_ok(root);
        uninstall_ok(root);
        assert_eq!(
            read_file(root, WORK),
            if malformed { OLD } else { ORIGINAL }
        );
        if malformed {
            assert_eq!(
                read_file(root, OWNERSHIP_FILE),
                "malformed provenance retained verbatim\n"
            );
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        }
    }
}

#[test]
fn m10_partial_file_operation_failure_preserves_failed_path_then_retries() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, FIRST, ORIGIN_CREATED, None);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    fs::create_dir(root.join(format!("{WORK}.tmp-zharness"))).unwrap();
    let mut out = String::new();
    assert!(run_update(
        UpdateOptions {
            root: root.into(),
            version: "new".into(),
            force: false
        },
        &mut out
    )
    .is_err());
    assert!(
        !root.join(FIRST).exists(),
        "first disposition should have succeeded: {out}"
    );
    assert_eq!(read_file(root, WORK), OLD);
    assert!(load_base(root).unwrap().contains_key(WORK));
    assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
    assert!(!out
        .lines()
        .any(|l| l.starts_with("restored") && l.contains(WORK)));
    fs::remove_dir(root.join(format!("{WORK}.tmp-zharness"))).unwrap();
    update_ok(root);
    assert_eq!(read_file(root, WORK), ORIGINAL);
    assert_relinquished(root, WORK);
    update_ok(root);
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), ORIGINAL);
}

#[test]
fn m10_metadata_failure_after_restore_keeps_backup_and_retry_finishes() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    fs::create_dir(root.join(format!("{MANIFEST_FILE}.tmp-zharness"))).unwrap();
    let mut out = String::new();
    assert!(run_update(
        UpdateOptions {
            root: root.into(),
            version: "new".into(),
            force: false
        },
        &mut out
    )
    .is_err());
    assert_eq!(read_file(root, WORK), ORIGINAL);
    assert!(load_base(root).unwrap().contains_key(WORK));
    assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
    assert_eq!(
        Ownership::load(root).get(KIND_FILE, WORK),
        ORIGIN_PREEXISTING
    );
    fs::remove_dir(root.join(format!("{MANIFEST_FILE}.tmp-zharness"))).unwrap();
    update_ok(root);
    assert_relinquished(root, WORK);
    assert!(find_original(root, WORK).is_none());
    assert!(!Ownership::load(root).has(KIND_FILE, WORK));
    update_ok(root);
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), ORIGINAL);
}

#[cfg(unix)]
#[test]
fn m11_retirement_directory_symlink_keeps_external_projection() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_CREATED, None);
    let outside = temp_repo();
    fs::rename(
        root.join("docs/playbooks"),
        outside.path().join("playbooks"),
    )
    .unwrap();
    std::os::unix::fs::symlink(
        outside.path().join("playbooks"),
        root.join("docs/playbooks"),
    )
    .unwrap();
    for _ in 0..2 {
        let out = update_ok(root);
        assert_eq!(read_file(root, WORK), OLD, "{out}");
        assert!(out.contains("symlink"));
    }
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), OLD);
    assert_eq!(
        fs::read_to_string(outside.path().join("playbooks/work.md")).unwrap(),
        OLD
    );
}

#[cfg(unix)]
#[test]
fn m11_retirement_leaf_symlink_never_changes_referent() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    let outside = temp_repo();
    fs::write(outside.path().join("target"), OLD).unwrap();
    fs::remove_file(root.join(WORK)).unwrap();
    std::os::unix::fs::symlink(outside.path().join("target"), root.join(WORK)).unwrap();
    update_ok(root);
    assert_eq!(
        fs::read_to_string(outside.path().join("target")).unwrap(),
        OLD
    );
    assert_eq!(read_file(root, WORK), ORIGINAL);
    assert!(!fs::symlink_metadata(root.join(WORK))
        .unwrap()
        .file_type()
        .is_symlink());
    update_ok(root);
    uninstall_ok(root);
    assert_eq!(
        fs::read_to_string(outside.path().join("target")).unwrap(),
        OLD
    );
    assert_eq!(read_file(root, WORK), ORIGINAL);
}

#[test]
fn m09_mixed_malformed_and_conflicting_provenance_preserved_verbatim() {
    let _env = IsolatedEnv::new();
    for damaged in [
        format!("file\tunrelated/path\tpreexisting\nfile\t{WORK}\nopaque provenance\n"),
        format!("file\tunrelated/path\tpreexisting\nfile\t{WORK}\tpreexisting\nfile\t{WORK}\tcreated_by_installer\n"),
        format!("file\tunrelated/path\tpreexisting\nfile\t{WORK}\tinvalid-origin\n"),
    ] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        write_file(root, OWNERSHIP_FILE, &damaged);
        for _ in 0..2 {
            update_ok(root);
            assert_eq!(read_file(root, WORK), OLD);
            assert_eq!(read_file(root, OWNERSHIP_FILE), damaged);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_relinquished(root, WORK);
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
        assert_eq!(read_file(root, OWNERSHIP_FILE), damaged);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
    }
}

#[test]
fn m09_conflicting_original_candidates_are_both_preserved() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    let other = format!("{ORIGINAL_DIR}/{}.orig", legacy_safe_path(WORK));
    write_file(root, &other, "different original\n");
    for _ in 0..2 {
        update_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
        assert_relinquished(root, WORK);
    }
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), OLD);
    assert_eq!(read_file(root, &other), "different original\n");
    assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
}

#[test]
fn m10_unlink_then_manifest_failure_retains_hash_ledger_blob_for_retry() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_CREATED, None);
    let blob = format!("{LEGACY_UPSTREAM_DIR}/old-proof");
    write_file(root, &blob, OLD);
    fs::create_dir(root.join(format!("{MANIFEST_FILE}.tmp-zharness"))).unwrap();
    let mut out = String::new();
    assert!(run_update(
        UpdateOptions {
            root: root.into(),
            version: "new".into(),
            force: false
        },
        &mut out
    )
    .is_err());
    assert!(!root.join(WORK).exists());
    assert!(load_base(root).unwrap().contains_key(WORK));
    assert_eq!(Ownership::load(root).get(KIND_FILE, WORK), ORIGIN_CREATED);
    assert_eq!(read_file(root, &blob), OLD);
    assert!(!out.contains("removed"));
    fs::remove_dir(root.join(format!("{MANIFEST_FILE}.tmp-zharness"))).unwrap();
    let out = update_ok(root);
    assert!(!out
        .lines()
        .any(|line| line.starts_with("removed") && line.contains(WORK)));
    assert_relinquished(root, WORK);
    assert!(!Ownership::load(root).has(KIND_FILE, WORK));
    uninstall_ok(root);
    assert!(!root.join(WORK).exists());
}

#[test]
fn m10_ownership_save_failure_keeps_original_and_retries_restoration_cleanup() {
    let _env = IsolatedEnv::new();
    for created in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(
            root,
            WORK,
            if created {
                ORIGIN_CREATED
            } else {
                ORIGIN_PREEXISTING
            },
            if created { None } else { Some(ORIGINAL) },
        );
        let blob = format!("{LEGACY_UPSTREAM_DIR}/old-proof");
        write_file(root, &blob, OLD);
        fs::create_dir(root.join(format!("{OWNERSHIP_FILE}.tmp-zharness"))).unwrap();
        let mut out = String::new();
        assert!(run_update(
            UpdateOptions {
                root: root.into(),
                version: "new".into(),
                force: false
            },
            &mut out
        )
        .is_err());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            if created {
                ORIGIN_CREATED
            } else {
                ORIGIN_PREEXISTING
            }
        );
        assert_eq!(read_file(root, &blob), OLD);
        assert_eq!(
            load_base(root).unwrap().get(WORK),
            Some(&sha(OLD.as_bytes()))
        );
        if !created {
            assert_eq!(read_file(root, WORK), ORIGINAL);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        } else {
            assert!(!root.join(WORK).exists());
        }
        fs::remove_dir(root.join(format!("{OWNERSHIP_FILE}.tmp-zharness"))).unwrap();
        update_ok(root);
        assert_relinquished(root, WORK);
        assert!(!Ownership::load(root).has(KIND_FILE, WORK));
        assert!(find_original(root, WORK).is_none());
        uninstall_ok(root);
        if !created {
            assert_eq!(read_file(root, WORK), ORIGINAL);
        }
    }
}

#[cfg(unix)]
#[test]
fn m11_docs_parent_and_original_symlink_shapes_keep_external_bytes() {
    let _env = IsolatedEnv::new();
    for shape in ["docs", "original-dir", "original-leaf"] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let outside = temp_repo();
        let external = outside.path().join("external");
        if shape == "docs" {
            fs::rename(root.join("docs"), &external).unwrap();
            std::os::unix::fs::symlink(&external, root.join("docs")).unwrap();
        } else if shape == "original-dir" {
            fs::rename(root.join(ORIGINAL_DIR), &external).unwrap();
            std::os::unix::fs::symlink(&external, root.join(ORIGINAL_DIR)).unwrap();
        } else {
            let original = find_original(root, WORK).unwrap();
            fs::rename(&original, &external).unwrap();
            std::os::unix::fs::symlink(&external, original).unwrap();
        }
        for _ in 0..2 {
            let out = update_ok(root);
            assert!(out.contains("symlink"), "{shape}: {out}");
            assert_eq!(read_file(root, WORK), OLD);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_relinquished(root, WORK);
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
        let external_bytes = if shape == "docs" {
            fs::read(external.join("playbooks/work.md"))
        } else if shape == "original-dir" {
            fs::read(external.join(format!("{}.orig", safe_path(WORK))))
        } else {
            fs::read(&external)
        }
        .unwrap();
        assert_eq!(
            external_bytes,
            if shape == "docs" {
                OLD.as_bytes()
            } else {
                ORIGINAL.as_bytes()
            }
        );
    }
}

#[cfg(unix)]
#[test]
fn m11_metadata_parent_link_skips_retirement_without_blocking_active_update() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    fs::remove_file(root.join(OWNERSHIP_FILE)).unwrap();
    let outside = temp_repo();
    let external = outside.path().join("base");
    fs::rename(root.join(BASE_DIR), &external).unwrap();
    std::os::unix::fs::symlink(&external, root.join(BASE_DIR)).unwrap();
    for _ in 0..2 {
        let out = update_ok(root);
        assert!(out.contains("symlink metadata"));
        assert_eq!(read_file(root, WORK), OLD);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert!(load_base(root).unwrap().contains_key(WORK));
        assert!(
            !external.join("ownership.tsv").exists(),
            "new retirement wrote through a metadata parent link"
        );
    }
}

#[test]
fn m09_unreadable_utf8_ledger_is_not_overwritten_by_migration() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_CREATED, None);
    fs::write(root.join(OWNERSHIP_FILE), [0xff, b'\n']).unwrap();
    let before = tree_snapshot(root);
    let mut out = String::new();
    assert!(run_update(
        UpdateOptions {
            root: root.into(),
            version: "new".into(),
            force: false
        },
        &mut out
    )
    .is_err());
    assert_eq!(tree_snapshot(root), before);
    assert!(install(root, "new", &mut out).is_err());
    assert_eq!(tree_snapshot(root), before);
    assert!(uninstall(root, &mut out).is_err());
    assert_eq!(tree_snapshot(root), before);
}

#[test]
fn m03_missing_base_already_original_never_authorizes_recovery_cleanup() {
    let _env = IsolatedEnv::new();
    for whole_manifest in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        write_file(root, WORK, ORIGINAL);
        if whole_manifest {
            fs::remove_file(root.join(MANIFEST_FILE)).unwrap();
        } else {
            let mut base = load_base(root).unwrap();
            base.remove(WORK);
            save_base(root, "old", &base).unwrap();
        }
        for _ in 0..2 {
            let out = update_ok(root);
            assert!(out.contains(WORK) && out.contains("unknown"));
            assert_eq!(read_file(root, WORK), ORIGINAL);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_eq!(
                Ownership::load(root).get(KIND_FILE, WORK),
                ORIGIN_PREEXISTING
            );
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
    }
}

#[test]
fn m10_backup_cleanup_failure_restores_base_ledger_and_each_original_for_retry() {
    let _env = IsolatedEnv::new();
    for fail_second in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let current = root
            .join(ORIGINAL_DIR)
            .join(format!("{}.orig", safe_path(WORK)));
        let old = root
            .join(ORIGINAL_DIR)
            .join(format!("{}.orig", legacy_safe_path(WORK)));
        fs::write(&old, ORIGINAL).unwrap();
        let prior_ledger = fs::read(root.join(OWNERSHIP_FILE)).unwrap();
        let prior_base = load_base(root).unwrap();
        let blob = format!("{LEGACY_UPSTREAM_DIR}/old-proof");
        write_file(root, &blob, OLD);
        super::retirement::fail_backup_unlink_once(if fail_second { &old } else { &current });
        let mut out = String::new();
        let error = run_update(
            UpdateOptions {
                root: root.into(),
                version: "new".into(),
                force: false,
            },
            &mut out,
        )
        .unwrap_err();
        assert!(error.contains("injected permission denied"), "{error}");
        assert!(out.contains("prior recovery evidence preserved"));
        assert!(!out
            .lines()
            .any(|line| line.starts_with("restored") && line.contains(WORK)));
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert_eq!(load_base(root).unwrap(), prior_base);
        assert_eq!(fs::read(root.join(OWNERSHIP_FILE)).unwrap(), prior_ledger);
        assert_eq!(fs::read(&current).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(fs::read(&old).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(read_file(root, &blob), OLD);
        update_ok(root);
        assert_relinquished(root, WORK);
        assert!(!Ownership::load(root).has(KIND_FILE, WORK));
        assert!(!current.exists() && !old.exists());
        update_ok(root);
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL);
    }
}

#[test]
fn m10_secondary_recovery_failure_is_explicit_and_retry_keeps_unknown_evidence() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    let current = find_original(root, WORK).unwrap();
    let old = root
        .join(ORIGINAL_DIR)
        .join(format!("{}.orig", legacy_safe_path(WORK)));
    fs::write(&old, ORIGINAL).unwrap();
    super::retirement::fail_backup_unlink_and_recovery_once(&old, root);
    let mut out = String::new();
    let error = run_update(
        UpdateOptions {
            root: root.into(),
            version: "new".into(),
            force: false,
        },
        &mut out,
    )
    .unwrap_err();
    assert!(error.contains("recovery also failed"), "{error}");
    assert!(out.contains("cleanup and recovery incomplete"));
    assert!(!out.contains("prior recovery evidence preserved"));
    assert!(!load_base(root).unwrap().contains_key(WORK));
    assert_eq!(
        Ownership::load(root).get(KIND_FILE, WORK),
        ORIGIN_PREEXISTING
    );
    assert_eq!(fs::read(&current).unwrap(), ORIGINAL.as_bytes());
    assert_eq!(fs::read(&old).unwrap(), ORIGINAL.as_bytes());
    fs::remove_dir(root.join(format!("{MANIFEST_FILE}.tmp-zharness"))).unwrap();
    for _ in 0..2 {
        let out = update_ok(root);
        assert!(out.contains(WORK) && out.contains("unknown"));
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert_eq!(fs::read(&current).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(fs::read(&old).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
    }
    uninstall_ok(root);
    assert_eq!(read_file(root, WORK), ORIGINAL);
    assert_eq!(fs::read(&current).unwrap(), ORIGINAL.as_bytes());
    assert_eq!(fs::read(&old).unwrap(), ORIGINAL.as_bytes());
}

#[test]
fn m07_failed_active_restore_or_read_retains_evidence_then_retry_succeeds() {
    let _env = IsolatedEnv::new();
    for read_failure in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        write_file(root, WORKFLOW_TARGET, ORIGINAL);
        install_ok(root);
        let before_base = load_base(root).unwrap();
        let before_ledger = fs::read(root.join(OWNERSHIP_FILE)).unwrap();
        let blocker = if read_failure {
            fs::remove_file(root.join(WORKFLOW_TARGET)).unwrap();
            root.join(WORKFLOW_TARGET)
        } else {
            root.join(format!("{WORKFLOW_TARGET}.tmp-zharness"))
        };
        fs::create_dir(&blocker).unwrap();
        let mut out = String::new();
        assert!(uninstall(root, &mut out).is_err());
        assert_eq!(load_base(root).unwrap(), before_base);
        assert_eq!(fs::read(root.join(OWNERSHIP_FILE)).unwrap(), before_ledger);
        assert_eq!(
            read_original(root, WORKFLOW_TARGET).unwrap(),
            ORIGINAL.as_bytes()
        );
        assert!(!out
            .lines()
            .any(|line| line.starts_with("restored") && line.contains(WORKFLOW_TARGET)));
        fs::remove_dir(&blocker).unwrap();
        if read_failure {
            write_file(
                root,
                WORKFLOW_TARGET,
                &crate::test_support::embedded_str("WORKFLOW.md"),
            );
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORKFLOW_TARGET), ORIGINAL);
    }
}

#[cfg(unix)]
#[test]
fn m11_install_and_uninstall_linked_recovery_preserve_external_payloads() {
    let _env = IsolatedEnv::new();
    for metadata_parent in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let outside = temp_repo();
        let target = outside.path().join("projection");
        fs::write(&target, OLD).unwrap();
        fs::remove_file(root.join(WORK)).unwrap();
        std::os::unix::fs::symlink(&target, root.join(WORK)).unwrap();
        let external = outside.path().join("recovery");
        let ledger_before = fs::read(root.join(OWNERSHIP_FILE)).unwrap();
        if metadata_parent {
            fs::rename(root.join(BASE_DIR), &external).unwrap();
            std::os::unix::fs::symlink(&external, root.join(BASE_DIR)).unwrap();
        } else {
            fs::rename(root.join(OWNERSHIP_FILE), &external).unwrap();
            std::os::unix::fs::symlink(&external, root.join(OWNERSHIP_FILE)).unwrap();
        }
        install_ok(root);
        uninstall_ok(root);
        assert_eq!(fs::read(&target).unwrap(), OLD.as_bytes());
        if metadata_parent {
            assert_eq!(read_file(root, WORK), OLD);
            assert_eq!(
                fs::read(
                    external
                        .join("original")
                        .join(format!("{}.orig", safe_path(WORK)))
                )
                .unwrap(),
                ORIGINAL.as_bytes()
            );
        } else {
            assert_eq!(read_file(root, WORK), ORIGINAL);
            assert_eq!(fs::read(&external).unwrap(), ledger_before);
        }
    }
}

#[test]
fn m02_initial_original_equality_is_not_restoration_proof() {
    let _env = IsolatedEnv::new();
    let repo = temp_repo();
    let root = repo.path();
    install_ok(root);
    legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
    write_file(root, WORK, ORIGINAL);
    for _ in 0..2 {
        let out = update_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
        assert!(
            out.lines()
                .any(|l| l.starts_with("KEPT") && l.contains(WORK)),
            "{out}"
        );
        assert_relinquished(root, WORK);
    }
    let out = uninstall_ok(root);
    assert_eq!(read_file(root, WORK), ORIGINAL);
    assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
    assert_eq!(
        Ownership::load(root).get(KIND_FILE, WORK),
        ORIGIN_PREEXISTING
    );
    assert!(
        out.lines()
            .any(|l| l.starts_with("KEPT") && l.contains(WORK)),
        "{out}"
    );
}

fn valid_restore_intent(root: &Path) -> (std::path::PathBuf, serde_json::Value) {
    let path = super::retirement::intent_path(root, WORK);
    let value = serde_json::json!({
        "version": 1, "path": WORK,
        "projection_sha256": sha(OLD.as_bytes()),
        "original_sha256": sha(ORIGINAL.as_bytes()),
        "origin": ORIGIN_PREEXISTING
    });
    (path, value)
}

#[test]
fn m02_invalid_or_stale_intent_never_licenses_cleanup_or_gets_overwritten() {
    let _env = IsolatedEnv::new();
    for mode in [
        "malformed",
        "path",
        "base",
        "original",
        "origin",
        "version",
        "unknown-field",
    ] {
        for local in [OLD, ORIGINAL] {
            let repo = temp_repo();
            let root = repo.path();
            install_ok(root);
            legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
            write_file(root, WORK, local);
            let (path, mut value) = valid_restore_intent(root);
            match mode {
                "path" => value["path"] = FIRST.into(),
                "base" => value["projection_sha256"] = sha(EDITED.as_bytes()).into(),
                "original" => value["original_sha256"] = sha(EDITED.as_bytes()).into(),
                "origin" => value["origin"] = ORIGIN_CREATED.into(),
                "version" => value["version"] = 2.into(),
                "unknown-field" => value["extra"] = true.into(),
                _ => {}
            }
            let raw = if mode == "malformed" {
                b"partial {".to_vec()
            } else {
                serde_json::to_vec(&value).unwrap()
            };
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(&path, &raw).unwrap();
            for _ in 0..2 {
                let out = update_ok(root);
                assert!(
                    out.contains("invalid or stale restoration intent"),
                    "{mode}: {out}"
                );
                assert_eq!(read_file(root, WORK), local);
                assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
                assert_eq!(
                    Ownership::load(root).get(KIND_FILE, WORK),
                    ORIGIN_PREEXISTING
                );
                assert_eq!(fs::read(&path).unwrap(), raw);
                assert_relinquished(root, WORK);
            }
            uninstall_ok(root);
            assert_eq!(read_file(root, WORK), local);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_eq!(
                Ownership::load(root).get(KIND_FILE, WORK),
                ORIGIN_PREEXISTING
            );
            assert_eq!(fs::read(&path).unwrap(), raw);
        }
    }
}

#[test]
fn m10_restore_intent_failure_and_interrupted_restore_retry() {
    let _env = IsolatedEnv::new();
    for stage in ["intent-write", "restore", "intent-cleanup"] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let prior_base = load_base(root).unwrap();
        let prior_own = fs::read(root.join(OWNERSHIP_FILE)).unwrap();
        let (intent_path, expected) = valid_restore_intent(root);
        let restore_tmp = root.join(format!("{WORK}.tmp-zharness"));
        match stage {
            "restore" => fs::create_dir(&restore_tmp).unwrap(),
            "intent-write" => super::retirement::fail_intent_once(&intent_path, false),
            _ => super::retirement::fail_intent_once(&intent_path, true),
        }
        let mut out = String::new();
        let error = run_update(
            UpdateOptions {
                root: root.into(),
                version: "new".into(),
                force: false,
            },
            &mut out,
        )
        .unwrap_err();
        assert!(
            !out.lines().any(
                |l| (l.starts_with("restored") || l.starts_with("completed")) && l.contains(WORK)
            ),
            "{out}"
        );
        assert_eq!(load_base(root).unwrap(), prior_base, "{stage}: {error}");
        assert_eq!(fs::read(root.join(OWNERSHIP_FILE)).unwrap(), prior_own);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            read_file(root, WORK),
            if stage == "intent-cleanup" {
                ORIGINAL
            } else {
                OLD
            }
        );
        if stage == "intent-write" {
            assert!(!intent_path.exists());
        } else {
            let saved: serde_json::Value =
                serde_json::from_slice(&fs::read(&intent_path).unwrap()).unwrap();
            assert_eq!(saved, expected);
        }
        if stage == "restore" {
            fs::remove_dir(&restore_tmp).unwrap();
        }
        let out = update_ok(root);
        assert!(
            out.lines()
                .any(|l| l.starts_with(if stage == "intent-cleanup" {
                    "completed"
                } else {
                    "restored"
                }) && l.contains(WORK)),
            "{out}"
        );
        assert_eq!(read_file(root, WORK), ORIGINAL);
        assert!(!intent_path.exists());
        assert!(find_original(root, WORK).is_none());
        assert!(!Ownership::load(root).has(KIND_FILE, WORK));
        assert_relinquished(root, WORK);
        update_ok(root);
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), ORIGINAL);
    }
}

#[cfg(unix)]
#[test]
fn m11_restore_intent_leaf_and_parent_symlinks_preserve_external_bytes() {
    let _env = IsolatedEnv::new();
    for parent in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let outside = tempfile::tempdir().unwrap();
        let (intent_path, expected) = valid_restore_intent(root);
        let raw = serde_json::to_vec(&expected).unwrap();
        let external = outside.path().join(intent_path.file_name().unwrap());
        fs::write(&external, &raw).unwrap();
        if parent {
            std::os::unix::fs::symlink(outside.path(), intent_path.parent().unwrap()).unwrap();
        } else {
            fs::create_dir_all(intent_path.parent().unwrap()).unwrap();
            std::os::unix::fs::symlink(&external, &intent_path).unwrap();
        }
        for _ in 0..2 {
            let out = update_ok(root);
            assert!(out.contains("invalid or stale restoration intent"), "{out}");
            assert_eq!(read_file(root, WORK), OLD);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_eq!(fs::read(&external).unwrap(), raw);
            assert_eq!(
                Ownership::load(root).get(KIND_FILE, WORK),
                ORIGIN_PREEXISTING
            );
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), OLD);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(fs::read(&external).unwrap(), raw);
        assert!(fs::symlink_metadata(if parent {
            intent_path.parent().unwrap()
        } else {
            &intent_path
        })
        .unwrap()
        .file_type()
        .is_symlink());
    }
}

#[test]
fn m03_valid_intent_cannot_replace_missing_base_or_override_later_edits() {
    let _env = IsolatedEnv::new();
    for missing_base in [false, true] {
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        legacy(root, WORK, ORIGIN_PREEXISTING, Some(ORIGINAL));
        let local = if missing_base { ORIGINAL } else { EDITED };
        write_file(root, WORK, local);
        let (path, value) = valid_restore_intent(root);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        let raw = serde_json::to_vec(&value).unwrap();
        fs::write(&path, &raw).unwrap();
        if missing_base {
            let mut base = load_base(root).unwrap();
            base.remove(WORK);
            save_base(root, "old", &base).unwrap();
        }
        for _ in 0..2 {
            let out = update_ok(root);
            assert!(
                out.lines()
                    .any(|l| l.starts_with("KEPT") && l.contains(WORK)),
                "{out}"
            );
            assert_eq!(read_file(root, WORK), local);
            assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
            assert_eq!(
                Ownership::load(root).get(KIND_FILE, WORK),
                ORIGIN_PREEXISTING
            );
            assert_eq!(fs::read(&path).unwrap(), raw);
            assert_relinquished(root, WORK);
        }
        uninstall_ok(root);
        assert_eq!(read_file(root, WORK), local);
        assert_eq!(read_original(root, WORK).unwrap(), ORIGINAL.as_bytes());
        assert_eq!(
            Ownership::load(root).get(KIND_FILE, WORK),
            ORIGIN_PREEXISTING
        );
        assert_eq!(fs::read(&path).unwrap(), raw);
    }
}
