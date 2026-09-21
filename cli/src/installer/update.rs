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
                "\nMove local text outside the markers, or rerun with --force to replace the block. Nothing was written.",
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
