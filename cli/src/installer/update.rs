//! `update`: refresh the managed set (ADR 0011).
//!
//! Playbooks and WORKFLOW.md are overwritten, `docs/PROJECT.md` is written
//! only when absent, and the AGENTS block is replaced between its markers
//! unless it was edited by hand since the last write. A single active plan in
//! the older 9-section format is migrated. Every refusal is decided before the
//! first write, so a refused update leaves every file as it was.

use std::fs;
use std::path::{Path, PathBuf};

use super::migrate::{self, PlanMigration};
use super::{
    agents_block_of, all_targets, apply_agents_block, canonical_agents_block, ensure_lines,
    gitignore_wants, load_base, registry, save_base, sha, src_bytes, write_file_atomic, Target,
    AGENTS_TARGET, GITIGNORE_TARGET, LEGACY_CONFLICTS_FILE, LEGACY_STASH_DIR,
};

/// Update invocation flags.
pub struct UpdateOptions {
    pub root: PathBuf,
    pub version: String,
    /// Replace a hand-edited AGENTS block instead of refusing.
    pub force: bool,
}

pub fn run_update(o: UpdateOptions, out: &mut String) -> Result<(), String> {
    let root = o.root.as_path();
    let targets = all_targets();
    let mut base_files = load_base(root)?;

    if root.join(LEGACY_CONFLICTS_FILE).exists() {
        out.push_str(&format!(
            "refused    a pre-0011 update stopped with unresolved conflicts ({LEGACY_CONFLICTS_FILE})\n\n"
        ));
        out.push_str(&format!(
            "Resolve the conflict markers in the files it lists, delete {LEGACY_CONFLICTS_FILE} and {LEGACY_STASH_DIR}/, then rerun. Nothing was written.\n"
        ));
        return Err(format!(
            "unresolved pre-0011 update conflicts in {LEGACY_CONFLICTS_FILE}"
        ));
    }

    let block_b = src_bytes(&Target {
        src: AGENTS_TARGET.to_string(),
        dst: AGENTS_TARGET.to_string(),
        once: false,
    })?;
    let want = canonical_agents_block(&String::from_utf8_lossy(&block_b));
    let ap = root.join(AGENTS_TARGET);
    let agents = fs::read(&ap).unwrap_or_default();
    let agents_readable = ap.exists();
    // A checkout with autocrlf is not a hand edit.
    let cur = agents_block_of(&String::from_utf8_lossy(&agents)).map(|c| c.replace("\r\n", "\n"));
    let tracked = base_files.get(AGENTS_TARGET).cloned();
    if let (Some(cur), Some(rec)) = (cur.as_ref(), tracked.as_ref()) {
        if !o.force && *cur != want && sha(cur.as_bytes()) != *rec {
            out.push_str(&format!(
                "refused    {AGENTS_TARGET}: the marked block was edited since zharness last wrote it\n\n"
            ));
            out.push_str(&unified_diff(
                &format!("{AGENTS_TARGET} (on disk)"),
                &format!("{AGENTS_TARGET} (this zharness)"),
                cur,
                &want,
            ));
            out.push_str(
                "\nMove local text outside the markers, or rerun with --force to replace the block. Nothing was written.\n",
            );
            return Err(format!(
                "{AGENTS_TARGET} block edited by hand; rerun with --force to replace it"
            ));
        }
    }

    let mig: PlanMigration = match migrate::prepare_plan_migration(root) {
        Ok(m) => m,
        Err(e) => {
            out.push_str(&format!(
                "refused    plan migration: {e}. Nothing was written.\n"
            ));
            return Err(e);
        }
    };

    let mut planned: Vec<(String, String)> = Vec::new();
    for t in &targets {
        let up = src_bytes(t)?;
        let dst_p = root.join(&t.dst);
        let mut note = "refreshed";
        if t.once {
            if dst_p.exists() {
                continue; // project-owned once written (ADR 0011)
            }
            note = "installed";
        }
        write_file_atomic(&dst_p, &up)?;
        base_files.insert(t.dst.clone(), sha(&up));
        planned.push((t.dst.clone(), note.to_string()));
    }

    if agents_readable {
        let (next, changed) = apply_agents_block(
            &String::from_utf8_lossy(&agents),
            &String::from_utf8_lossy(&block_b),
        );
        if changed {
            write_file_atomic(&ap, next.as_bytes())?;
            let note = if cur.is_none() {
                "block appended"
            } else {
                "block refreshed"
            };
            planned.push((AGENTS_TARGET.to_string(), note.to_string()));
        }
        base_files.insert(AGENTS_TARGET.to_string(), sha(want.as_bytes()));
    }

    if !mig.path.is_empty() {
        migrate::apply_plan_migration(root, &mig)?;
        planned.push((mig.path.clone(), "migrated".to_string()));
    }
    if !mig.notice.is_empty() {
        out.push_str(&format!("notice     {}\n", mig.notice));
    }

    if let Ok(Some(note)) = reconcile_gitignore(root, &gitignore_wants()) {
        planned.push((GITIGNORE_TARGET.to_string(), note));
    }

    save_base(root, &o.version, &base_files)?;
    registry::register(root, out);

    planned.sort_by(|a, b| a.0.cmp(&b.0));
    for (name, note) in &planned {
        out.push_str(&format!("{note:<14} {name}\n"));
    }
    out.push_str("update complete.\n");
    Ok(())
}

/// Renders `a` against `b` as a single hunk around their differing middle with
/// up to three lines of context. The AGENTS block is short, so one hunk stays
/// readable and needs no line-matching algorithm.
fn unified_diff(from: &str, to: &str, a: &str, b: &str) -> String {
    let al: Vec<&str> = a.split('\n').collect();
    let bl: Vec<&str> = b.split('\n').collect();
    let mut p = 0;
    while p < al.len() && p < bl.len() && al[p] == bl[p] {
        p += 1;
    }
    let mut s = 0;
    while s < al.len().saturating_sub(p)
        && s < bl.len().saturating_sub(p)
        && al[al.len() - 1 - s] == bl[bl.len() - 1 - s]
    {
        s += 1;
    }
    let lo = p.saturating_sub(3);
    let tail = s.min(3);
    let mut w = String::new();
    w.push_str(&format!(
        "--- {from}\n+++ {to}\n@@ -{},{} +{},{} @@\n",
        lo + 1,
        al.len() - s + tail - lo,
        lo + 1,
        bl.len() - s + tail - lo
    ));
    for l in &al[lo..p] {
        w.push_str(&format!(" {l}\n"));
    }
    for l in &al[p..al.len() - s] {
        w.push_str(&format!("-{l}\n"));
    }
    for l in &bl[p..bl.len() - s] {
        w.push_str(&format!("+{l}\n"));
    }
    for l in &al[al.len() - s..al.len() - s + tail] {
        w.push_str(&format!(" {l}\n"));
    }
    w
}

/// Re-assert the managed ignore lines. `None` means nothing changed.
fn reconcile_gitignore(root: &Path, wants: &[&str]) -> Result<Option<String>, String> {
    let gp = root.join(GITIGNORE_TARGET);
    let now = match fs::read_to_string(&gp) {
        Ok(s) => s,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(e) => return Err(format!("read {}: {e}", gp.display())),
    };
    let body = ensure_lines(&now, wants);
    if body == now {
        return Ok(None);
    }
    write_file_atomic(&gp, body.as_bytes())?;
    Ok(Some("+ ignore entries re-asserted".to_string()))
}

#[cfg(test)]
mod tests {
    use std::fs;

    use super::super::check::check;
    use super::super::ownership::OWNERSHIP_FILE;
    use super::super::{
        canonical_agents_block, load_base, save_base, sha, uninstall, AGENTS_TARGET,
        LEGACY_STASH_DIR, LEGACY_UPSTREAM_DIR, PROJECT_TARGET, WORKFLOW_TARGET, ZHRNESS_DIR,
    };
    use super::*;
    use crate::test_support::{
        embedded_str, install_ok, read_file, temp_repo, tree_snapshot, update_ok, write_file,
        IsolatedEnv,
    };

    /// docs/PROJECT.md is write-once (ADR 0011): update scaffolds it when
    /// absent and never touches an existing one, whatever the template does.
    #[test]
    fn update_project_write_once() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);

        let shipped = read_file(root, PROJECT_TARGET);
        let custom = shipped.replace(
            "## Who is it for?",
            "## Who is it for?\n- project-owned answer",
        );
        assert_ne!(custom, shipped, "fixture did not answer the identity file");
        write_file(root, PROJECT_TARGET, &custom);
        update_ok(root);
        assert_eq!(
            read_file(root, PROJECT_TARGET),
            custom,
            "update changed an existing PROJECT.md"
        );

        fs::remove_file(root.join(PROJECT_TARGET)).unwrap();
        let out = update_ok(root);
        assert_eq!(
            read_file(root, PROJECT_TARGET),
            shipped,
            "absent PROJECT.md not scaffolded from the template"
        );
        assert!(
            out.contains("installed"),
            "expected an installed line:\n{out}"
        );
    }

    /// Playbooks and WORKFLOW.md are pure upstream mirrors: update always
    /// overwrites them with upstream bytes, ignoring any local edit and never
    /// producing a conflict.
    #[test]
    fn update_fresh_overwrite_playbooks_and_workflow_ignore_local_edits() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let upstream_wf = read_file(root, WORKFLOW_TARGET);
        let upstream_play = read_file(root, "docs/playbooks/work.md");

        write_file(
            root,
            WORKFLOW_TARGET,
            "totally different hand-edited content\n",
        );
        update_ok(root);
        assert_eq!(
            read_file(root, WORKFLOW_TARGET),
            upstream_wf,
            "expected fresh overwrite to win over local edit"
        );

        write_file(root, "docs/playbooks/work.md", "hand-edited playbook\n");
        update_ok(root);
        assert_eq!(
            read_file(root, "docs/playbooks/work.md"),
            upstream_play,
            "playbook local edit should be silently discarded"
        );
    }

    /// The AGENTS block is replaced between its markers unless it was edited
    /// since zharness last wrote it (ADR 0011). These tests pin the
    /// replace-and-keep-prose branches; the refusal half has no end-to-end
    /// test since the golden fixtures were removed.
    #[test]
    fn update_agents_block_untouched_is_replaced_and_prose_kept() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);

        // "zharness last wrote an older block": the on-disk block differs from
        // the embedded one and the recorded base names it, which is the state
        // the guard accepts as untouched.
        let older = canonical_agents_block(&format!(
            "{}\nolder-upstream-line",
            embedded_str(AGENTS_TARGET)
        ));
        write_file(
            root,
            AGENTS_TARGET,
            &format!("consumer prose above\n\n{older}\nconsumer prose below\n"),
        );
        let mut base = load_base(root).unwrap();
        base.insert(AGENTS_TARGET.to_string(), sha(older.as_bytes()));
        save_base(root, "test", &base).unwrap();

        update_ok(root);
        let got = read_file(root, AGENTS_TARGET);
        assert!(
            got.contains("no parallel control-plane state"),
            "block not refreshed:\n{got}"
        );
        assert!(
            !got.contains("older-upstream-line"),
            "old block survived:\n{got}"
        );
        assert!(
            got.starts_with("consumer prose above\n") && got.ends_with("consumer prose below\n"),
            "prose outside the markers changed:\n{got}"
        );

        update_ok(root);
        assert_eq!(
            read_file(root, AGENTS_TARGET),
            got,
            "update is not idempotent for the AGENTS block"
        );
    }

    #[test]
    fn update_agents_block_without_recorded_hash_is_accepted_and_recorded() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let edited = read_file(root, AGENTS_TARGET).replace(
            "no parallel control-plane state",
            "consumer rewrote this line",
        );
        write_file(root, AGENTS_TARGET, &edited);
        let mut base = load_base(root).unwrap();
        base.remove(AGENTS_TARGET);
        save_base(root, "legacy", &base).unwrap();

        update_ok(root);
        let base = load_base(root).unwrap();
        let want = sha(canonical_agents_block(&embedded_str(AGENTS_TARGET)).as_bytes());
        assert_eq!(
            base.get(AGENTS_TARGET),
            Some(&want),
            "block hash not recorded: {:?}",
            base.get(AGENTS_TARGET)
        );
    }

    #[test]
    fn update_agents_block_crlf_checkout_is_not_a_hand_edit() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        write_file(
            root,
            AGENTS_TARGET,
            &read_file(root, AGENTS_TARGET).replace('\n', "\r\n"),
        );

        let out = update_ok(root);
        assert!(
            !out.contains("refused"),
            "a CRLF checkout of an untouched block must not be refused:\n{out}"
        );
        let want = canonical_agents_block(&embedded_str(AGENTS_TARGET));
        assert_eq!(
            agents_block_of(&read_file(root, AGENTS_TARGET)).as_deref(),
            Some(want.as_str()),
            "block not refreshed to the canonical bytes"
        );
    }

    /// A pre-ADR 0011 installation may carry an unresolved conflict list. Its
    /// presence makes update refuse before any write, and check report drift.
    #[test]
    fn update_legacy_conflicts_refuses_and_check_reports() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        write_file(root, LEGACY_CONFLICTS_FILE, "[\"docs/PROJECT.md\"]\n");
        write_file(root, WORKFLOW_TARGET, "stale workflow\n");
        let before = tree_snapshot(root);

        let mut out = String::new();
        let err = run_update(
            UpdateOptions {
                root: root.to_path_buf(),
                version: "t".to_string(),
                force: true,
            },
            &mut out,
        );
        assert!(
            err.is_err(),
            "expected refusal while a pre-0011 conflict is unresolved"
        );
        assert_eq!(tree_snapshot(root), before, "refused update wrote files");
        assert!(
            out.contains(LEGACY_CONFLICTS_FILE),
            "refusal must name the conflict list:\n{out}"
        );

        let drift = check(root).unwrap();
        assert!(
            drift
                .join("\n")
                .contains(&format!("conflict {LEGACY_CONFLICTS_FILE}")),
            "check must report the unresolved conflict: {drift:?}"
        );
    }

    /// A pre-ADR 0011 installation carries content-addressed blobs under
    /// `.zharness/base/upstream/` and may carry an update stash. Its manifest
    /// already holds the hashes update needs, so update deletes the blobs,
    /// keeps the ledger, and leaves the stash for uninstall.
    #[test]
    fn update_legacy_artifacts_blobs_dropped_ledger_kept() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        write_file(
            root,
            &format!("{LEGACY_UPSTREAM_DIR}/deadbeef.bin"),
            "old blob\n",
        );
        write_file(
            root,
            &format!("{LEGACY_STASH_DIR}/stash.tsv"),
            "docs/PROJECT.md\tx.bin\t1\n",
        );
        let ledger = read_file(root, OWNERSHIP_FILE);

        update_ok(root);
        assert!(
            !root.join(LEGACY_UPSTREAM_DIR).exists(),
            "update kept the legacy blob store"
        );
        assert_eq!(
            read_file(root, OWNERSHIP_FILE),
            ledger,
            "update changed the ownership ledger"
        );
        assert!(
            root.join(LEGACY_STASH_DIR).exists(),
            "update deleted a legacy stash; only uninstall may"
        );

        let mut out = String::new();
        uninstall(root, &mut out).unwrap_or_else(|e| panic!("uninstall: {e}\n{out}"));
        assert!(
            !root.join(ZHRNESS_DIR).exists(),
            "uninstall left .zharness behind"
        );
    }

    /// The shipped templates/project.identity.md exactly as it stood before
    /// the gate-slot rewrite (commit aba7057). It is the base a consumer
    /// installed against, so the update under test replays the real upgrade
    /// path rather than a synthetic one.
    const IDENTITY_PRE_EDIT: &str = r#"# PROJECT — identity (answer inline; this is the single forced write step at
# brainstorm lock; keep the whole file at or under 50 lines)

## What is this project?
- <one sentence: what the product IS>

## Who is it for?
- <primary users/teams>

## Non-goals
- <explicitly excluded scope>

## How do we run the tests?
- `<exact verification command(s)>`

## Architecture in one breath
- runtime shape: <...>
- where state lives: <...>
- entrypoints: <...>

## What are we working on right now?
- plan: docs/plans/active/<slug>.md (<status>)
"#;

    /// A consumer who answered the pre-gate-slot identity template keeps the
    /// file byte for byte on update; the new question surfaces through
    /// `update --check` as a missing heading instead of a merge conflict
    /// (ADR 0011).
    #[test]
    fn update_identity_template_change_keeps_project_check_names_heading() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let shipped = embedded_str(super::super::PROJECT_TEMPLATE);
        assert!(
            shipped.contains("## What are the gate commands?"),
            "shipped template lacks the gate section under test:\n{shipped}"
        );

        let filled =
            IDENTITY_PRE_EDIT.replace("- `<exact verification command(s)>`", "- `pnpm test`");
        assert_ne!(
            filled, IDENTITY_PRE_EDIT,
            "fixture did not fill the tests answer"
        );
        write_file(root, PROJECT_TARGET, &filled);

        update_ok(root);
        assert_eq!(
            read_file(root, PROJECT_TARGET),
            filled,
            "update changed the filled identity file"
        );
        let drift = check(root).unwrap();
        assert!(
            drift
                .join("\n")
                .contains("lacks \"## What are the gate commands?\""),
            "check does not name the new heading:\n{}",
            drift.join("\n")
        );
    }
}
