//! `uninstall`: remove exactly what the ownership ledger attributes to the
//! installer (ADR 0008).
//!
//! A managed file whose current bytes differ from both its recorded base and
//! the embedded upstream is locally authored work and is left in place with a
//! warning; so is anything whose provenance is unknown, because a leftover
//! file costs the consumer nothing they cannot delete while a wrong guess
//! costs them work they cannot recover.

use std::fs;
use std::path::Path;

use super::ownership::{self, Ownership, KIND_FILE, ORIGIN_CREATED, ORIGIN_PREEXISTING};
use super::{
    agents_span, all_targets, load_base, read_original, registry, remove_dir_if_empty, sha,
    write_file_atomic, AGENTS_CREATED_HEADER, AGENTS_TARGET, BASE_DIR, GITIGNORE_TARGET,
    LEGACY_CONFLICTS_FILE, LEGACY_STASH_DIR, LEGACY_UPSTREAM_DIR, MANIFEST_FILE, ORIGINAL_DIR,
    ZHRNESS_DIR,
};

pub fn uninstall(root: &Path, out: &mut String) -> Result<(), String> {
    let targets = all_targets();
    let base_files = load_base(root)?;
    let own = ownership::load_ownership_for(root, &targets, &base_files);

    let mut removed = 0usize;
    let mut kept: Vec<String> = Vec::new();
    for t in &targets {
        let base = base_files.get(&t.dst).map(String::as_str);
        remove_managed_file(
            root,
            &t.dst,
            base,
            own.get(KIND_FILE, &t.dst),
            &mut removed,
            &mut kept,
            out,
        );
    }
    remove_agents_block(root, &own, &mut removed, &mut kept, out);

    let gi_now = fs::read(root.join(GITIGNORE_TARGET)).unwrap_or_default();
    let cleaned = drop_lines(&gi_now, &own.owned_gitignore_lines());
    if cleaned != gi_now {
        let gi_origin = own.get(KIND_FILE, GITIGNORE_TARGET);
        if String::from_utf8_lossy(&cleaned).trim().is_empty() && gi_origin == ORIGIN_CREATED {
            let _ = fs::remove_file(root.join(GITIGNORE_TARGET));
            out.push_str(&format!("removed   {GITIGNORE_TARGET}\n"));
            removed += 1;
        } else {
            let _ = write_file_atomic(&root.join(GITIGNORE_TARGET), &cleaned);
            out.push_str(&format!(
                "restored  {GITIGNORE_TARGET} (zharness entries removed)\n"
            ));
        }
    }

    let _ = fs::remove_dir_all(root.join(LEGACY_STASH_DIR));
    let _ = fs::remove_file(root.join(LEGACY_CONFLICTS_FILE));
    let _ = fs::remove_file(root.join(MANIFEST_FILE));
    let _ = fs::remove_file(root.join(ownership::OWNERSHIP_FILE));
    let _ = fs::remove_dir_all(root.join(LEGACY_UPSTREAM_DIR));
    let _ = fs::remove_dir_all(root.join(ORIGINAL_DIR));
    let _ = fs::remove_dir_all(root.join(BASE_DIR));
    remove_dir_if_empty(&root.join(ZHRNESS_DIR));
    registry::unregister(root, out);

    // Report what happened, never a blanket guarantee the run did not enforce:
    // a gate that overstates its own coverage is worse than one that says
    // nothing (ADR 0008).
    if kept.is_empty() {
        out.push_str(&format!(
            "uninstall complete — {removed} managed file(s) removed; nothing else was touched.\n"
        ));
        return Ok(());
    }
    out.push_str(&format!(
        "uninstall complete — {removed} managed file(s) removed, {} kept:\n",
        kept.len()
    ));
    for k in &kept {
        out.push_str(&format!("  {k}\n"));
    }
    Ok(())
}

fn drop_lines(blob: &[u8], lines: &[String]) -> Vec<u8> {
    let mut s = String::from_utf8_lossy(blob).into_owned();
    for l in lines {
        s = drop_line(&s, l);
    }
    s.into_bytes()
}

fn drop_line(blob: &str, want: &str) -> String {
    let keep: Vec<&str> = blob
        .split('\n')
        .filter(|ln| ln.trim() != want.trim())
        .collect();
    let mut out = keep.join("\n");
    while out.contains("\n\n\n") {
        out = out.replace("\n\n\n", "\n\n");
    }
    out = format!("{}\n", out.trim_end_matches('\n'));
    if out == "\n" {
        return String::new();
    }
    out
}

/// Compares against the recorded base (the sha256 of the bytes zharness last
/// wrote), not the live embedded bytes. The ledger decides whether the file
/// may be deleted at all.
fn remove_managed_file(
    root: &Path,
    rel: &str,
    base: Option<&str>,
    origin: String,
    removed: &mut usize,
    kept: &mut Vec<String>,
    out: &mut String,
) {
    let dst_p = root.join(rel);
    let local = match fs::read(&dst_p) {
        Ok(b) => b,
        Err(_) => return,
    };
    let orig = read_original(root, rel);
    let keep = |reason: &str, kept: &mut Vec<String>, out: &mut String| {
        out.push_str(&format!("KEPT      {rel} ({reason})\n"));
        kept.push(format!("{rel} — {reason}"));
    };
    match base {
        // No recorded base at all: the file cannot be compared to anything, so
        // it is not "locally modified" — it is unattributable. Say that.
        None => keep(
            "no recorded base; provenance unknown, delete manually if intended",
            kept,
            out,
        ),
        Some(base) if sha(&local) == base => match (origin.as_str(), orig.as_ref()) {
            (ORIGIN_PREEXISTING, Some(o)) => {
                let _ = fs::write(&dst_p, o);
                out.push_str(&format!("restored  {rel} (pre-install original)\n"));
            }
            (ORIGIN_CREATED, _) => {
                let _ = fs::remove_file(&dst_p);
                out.push_str(&format!("removed   {rel}\n"));
                *removed += 1;
            }
            // preexisting with a lost original, or no recorded origin at all:
            // unknown provenance is never resolved by deleting.
            _ => keep("provenance unknown; delete manually if intended", kept, out),
        },
        Some(_) => match &orig {
            Some(o) if *o == local => {
                out.push_str(&format!(
                    "restored  {rel} (already at pre-install original)\n"
                ));
            }
            _ => keep("locally modified; delete manually if intended", kept, out),
        },
    }
    if let Some(dir) = dst_p.parent() {
        remove_dir_if_empty(dir);
    }
}

fn remove_agents_block(
    root: &Path,
    own: &Ownership,
    removed: &mut usize,
    kept: &mut Vec<String>,
    out: &mut String,
) {
    let ap = root.join(AGENTS_TARGET);
    let raw = match fs::read(&ap) {
        Ok(b) => b,
        Err(_) => return,
    };
    let content = String::from_utf8_lossy(&raw).into_owned();
    let Some((i, ej)) = agents_span(&content) else {
        return;
    };
    let mut end = ej;
    if end < content.len() && content.as_bytes()[end] == b'\n' {
        end += 1;
    }
    let outside = format!("{}{}", &content[..i], &content[end..]);
    let remainder = outside.trim();
    let origin = own.get(KIND_FILE, AGENTS_TARGET);

    // Creating the file does not confer ownership of every byte written into
    // it afterwards. Delete only when nothing but this installer's own
    // generated header remains.
    let generated_only = origin == ORIGIN_CREATED && remainder == AGENTS_CREATED_HEADER;

    if remainder.is_empty() {
        let _ = fs::remove_file(&ap);
        out.push_str(&format!(
            "removed   {AGENTS_TARGET} (nothing outside the block)\n"
        ));
        *removed += 1;
    } else if generated_only {
        let _ = fs::remove_file(&ap);
        out.push_str(&format!(
            "removed   {AGENTS_TARGET} (created by install; only generated boilerplate outside the block)\n"
        ));
        *removed += 1;
    } else {
        let body = format!("{}\n", outside.trim_end_matches('\n'));
        let _ = write_file_atomic(&ap, body.as_bytes());
        out.push_str(&format!(
            "unmarked  {AGENTS_TARGET} (block removed, surrounding text preserved)\n"
        ));
        kept.push(format!(
            "{AGENTS_TARGET} — text outside the block preserved"
        ));
    }
    if let Some(dir) = ap.parent() {
        remove_dir_if_empty(dir);
    }
}
