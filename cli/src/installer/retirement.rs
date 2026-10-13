//! Fixed legacy recovery inventory and conservative retirement (ADR 0015).
use std::collections::HashMap;
use std::fs;
use std::io::Write;

use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};

use super::ownership::{self, Ownership, KIND_FILE, ORIGIN_CREATED, ORIGIN_PREEXISTING};
use super::{
    all_targets, legacy_safe_path, safe_path, save_base_preserving_blobs, sha, write_file_atomic,
    Target, MANIFEST_FILE, ORIGINAL_DIR, PLAYBOOK_DIR_TGT,
};

const NAMES: [&str; 12] = [
    "brainstorm-grill.md",
    "brainstorm-lock.md",
    "brainstorm.md",
    "check-durable.md",
    "check-review.md",
    "check-validation.md",
    "check.md",
    "handoff.md",
    "to-plan.md",
    "watzup.md",
    "work-full.md",
    "work.md",
];

pub(super) fn is_retired(path: &str) -> bool {
    NAMES
        .iter()
        .any(|name| path == format!("{PLAYBOOK_DIR_TGT}/{name}"))
}

pub(super) fn recovery_targets() -> Vec<Target> {
    let mut targets = all_targets();
    targets.extend(NAMES.iter().map(|name| Target {
        src: format!("playbooks/{name}"),
        dst: format!("{PLAYBOOK_DIR_TGT}/{name}"),
        once: false,
    }));
    targets
}

pub(super) fn load_ownership(
    root: &Path,
    base: &HashMap<String, String>,
) -> Result<Ownership, String> {
    match fs::read_to_string(root.join(ownership::OWNERSHIP_FILE)) {
        Ok(_) => {}
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(e) => return Err(format!("read ownership: {e}")),
    }
    Ok(ownership::load_ownership_for(
        root,
        &recovery_targets(),
        base,
    ))
}

/// Parents only: the destination leaf is replaced/unlinked atomically. This
/// boundary applies to retirement, not ADR 0013's existing managed writers.
pub(super) fn symlink_parent(root: &Path, path: &Path) -> bool {
    let Ok(rel) = path.strip_prefix(root) else {
        return true;
    };
    let mut p = root.to_path_buf();
    let Some(parent) = rel.parent() else {
        return false;
    };
    for part in parent.components() {
        p.push(part);
        if fs::symlink_metadata(&p).is_ok_and(|m| m.file_type().is_symlink()) {
            return true;
        }
    }
    false
}

fn original_paths(root: &Path, rel: &str) -> Vec<PathBuf> {
    [safe_path(rel), legacy_safe_path(rel)]
        .into_iter()
        .map(|name| root.join(ORIGINAL_DIR).join(format!("{name}.orig")))
        .filter(|p| fs::symlink_metadata(p).is_ok())
        .collect()
}

// Recovery metadata only: recorded before an eligible original restoration,
// never inferred from equality after the fact. Separate from active base/ledger.
const INTENT_DIR: &str = ".zharness/base/retirement-intent";

#[derive(Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
struct RestoreIntent {
    version: u8,
    path: String,
    projection_sha256: String,
    original_sha256: String,
    origin: String,
}

pub(super) fn intent_path(root: &Path, rel: &str) -> PathBuf {
    root.join(INTENT_DIR)
        .join(format!("{}.json", safe_path(rel)))
}

fn expected_intent(rel: &str, base: &str, original: &[u8]) -> RestoreIntent {
    RestoreIntent {
        version: 1,
        path: rel.to_owned(),
        projection_sha256: base.to_owned(),
        original_sha256: sha(original),
        origin: ORIGIN_PREEXISTING.to_owned(),
    }
}

// Missing differs from malformed/unreadable/linked. An existing invalid record
// is retained rather than overwritten with newly manufactured proof.
fn load_intent(root: &Path, path: &Path) -> Result<Option<RestoreIntent>, ()> {
    if symlink_parent(root, path) {
        return Err(());
    }
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Ok(m) if m.is_file() && !m.file_type().is_symlink() => {
            serde_json::from_slice(&fs::read(path).map_err(|_| ())?)
                .map(Some)
                .map_err(|_| ())
        }
        _ => Err(()),
    }
}

fn record_intent(root: &Path, path: &Path, intent: &RestoreIntent) -> Result<(), String> {
    #[cfg(test)]
    fail_intent_operation(path, false)?;
    if symlink_parent(root, path) {
        return Err("record retirement intent: symlink parent".to_owned());
    }
    let parent = path.parent().unwrap();
    fs::create_dir_all(parent).map_err(|e| format!("mkdir retirement intent: {e}"))?;
    let bytes = serde_json::to_vec(intent).map_err(|e| format!("encode retirement intent: {e}"))?;
    // create_new refuses a planted leaf (including a symlink). Flush proof
    // before changing consumer bytes; a partial intent remains invalid/kept.
    let mut file = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(path)
        .map_err(|e| format!("create retirement intent: {e}"))?;
    file.write_all(&bytes)
        .and_then(|_| file.sync_all())
        .map_err(|e| format!("persist retirement intent: {e}"))?;
    #[cfg(unix)]
    {
        fs::File::open(parent)
            .and_then(|dir| dir.sync_all())
            .map_err(|e| format!("persist retirement intent directory: {e}"))?;
        // The intent directory may itself have just been created.
        fs::File::open(parent.parent().unwrap())
            .and_then(|dir| dir.sync_all())
            .map_err(|e| format!("persist retirement intent parent: {e}"))?;
    }
    Ok(())
}

/// Duplicate contradictory rows are not proof, regardless of parser order.
fn origin_is_ambiguous(root: &Path, rel: &str) -> bool {
    let Ok(raw) = fs::read_to_string(root.join(ownership::OWNERSHIP_FILE)) else {
        return false;
    };
    let mut origin = None;
    for line in raw.lines() {
        let parts: Vec<_> = line.split('\t').collect();
        if parts.len() >= 2 && parts[0] == KIND_FILE && parts[1] == rel {
            if parts.len() != 3
                || ![ORIGIN_CREATED, ORIGIN_PREEXISTING, "unknown"].contains(&parts[2])
            {
                return true;
            }
            if origin.is_some_and(|old| old != parts[2]) {
                return true;
            }
            origin = Some(parts[2]);
        }
    }
    false
}

#[cfg(test)]
thread_local! {
    static FAIL_BACKUP_UNLINK: std::cell::RefCell<Option<(PathBuf, Option<PathBuf>)>> = const { std::cell::RefCell::new(None) };
}

#[cfg(test)]
pub(super) fn fail_backup_unlink_once(path: &Path) {
    FAIL_BACKUP_UNLINK.with(|slot| *slot.borrow_mut() = Some((path.to_owned(), None)));
}

#[cfg(test)]
pub(super) fn fail_backup_unlink_and_recovery_once(path: &Path, root: &Path) {
    let blocker = root.join(format!("{MANIFEST_FILE}.tmp-zharness"));
    FAIL_BACKUP_UNLINK.with(|slot| *slot.borrow_mut() = Some((path.to_owned(), Some(blocker))));
}

#[cfg(test)]
thread_local! {
    static FAIL_INTENT: std::cell::RefCell<Option<(PathBuf, bool)>> = const { std::cell::RefCell::new(None) };
}

#[cfg(test)]
pub(super) fn fail_intent_once(path: &Path, cleanup: bool) {
    FAIL_INTENT.with(|slot| *slot.borrow_mut() = Some((path.to_owned(), cleanup)));
}

#[cfg(test)]
fn fail_intent_operation(path: &Path, cleanup: bool) -> Result<(), String> {
    if FAIL_INTENT.with(|slot| {
        let mut slot = slot.borrow_mut();
        if slot
            .as_ref()
            .is_some_and(|(selected, c)| selected == path && *c == cleanup)
        {
            slot.take();
            true
        } else {
            false
        }
    }) {
        return Err(format!(
            "{} retirement intent: injected permission denied",
            if cleanup { "remove" } else { "create" }
        ));
    }
    Ok(())
}

fn remove_backup(path: &Path) -> Result<(), String> {
    #[cfg(test)]
    if let Some(blocker) = FAIL_BACKUP_UNLINK.with(|slot| {
        let mut slot = slot.borrow_mut();
        if slot.as_ref().is_some_and(|(selected, _)| selected == path) {
            slot.take().map(|(_, blocker)| blocker)
        } else {
            None
        }
    }) {
        if let Some(blocker) = blocker {
            fs::create_dir(&blocker).map_err(|e| format!("inject recovery blocker: {e}"))?;
        }
        return Err(format!(
            "remove restored backup {}: injected permission denied",
            path.display()
        ));
    }
    fs::remove_file(path).map_err(|e| format!("remove restored backup {}: {e}", path.display()))
}

/// A returned cleanup failure must not turn absent metadata into authority on
/// retry. Restore the original evidence; report any secondary recovery error
/// rather than claiming the transaction rolled back successfully.
fn rollback_cleanup(
    root: &Path,
    version: &str,
    prior_base: &HashMap<String, String>,
    prior_own: Option<&Ownership>,
    removed_originals: &[PathBuf],
    payload: Option<&Vec<u8>>,
) -> Vec<String> {
    let mut failures = Vec::new();
    for path in removed_originals {
        if symlink_parent(root, path) {
            failures.push(format!("restore backup {}: symlink parent", path.display()));
        } else if let Some(bytes) = payload {
            if let Err(e) = write_file_atomic(path, bytes) {
                failures.push(e);
            }
        }
    }
    if let Some(own) = prior_own {
        if let Err(e) = own.save(root) {
            failures.push(e);
        }
    }
    if let Err(e) = save_base_preserving_blobs(root, version, prior_base) {
        failures.push(e);
    }
    failures
}

pub(super) fn retire(
    root: &Path,
    version: &str,
    base: &mut HashMap<String, String>,
    own: &mut Ownership,
    out: &mut String,
) -> Result<(), String> {
    // Preserve ADR 0013's existing active writers. Retirement itself does no
    // metadata operation through linked parents; leave the evidence in place.
    if symlink_parent(root, &root.join(MANIFEST_FILE))
        || symlink_parent(root, &root.join(ownership::OWNERSHIP_FILE))
    {
        out.push_str("KEPT      legacy procedures (symlink metadata parent; recovery unchanged)\n");
        return Ok(());
    }
    // A missing ledger is seeded from the combined inventory before any file
    // operation, and its evidence survives an interrupted retirement.
    if fs::symlink_metadata(root.join(ownership::OWNERSHIP_FILE))
        .is_err_and(|e| e.kind() == std::io::ErrorKind::NotFound)
        && !base.is_empty()
    {
        own.save(root)?;
    }
    for name in NAMES {
        let rel = format!("{PLAYBOOK_DIR_TGT}/{name}");
        let dst = root.join(&rel);
        let originals = original_paths(root, &rel);
        let intent_path = intent_path(root, &rel);
        let intent = load_intent(root, &intent_path);
        let present = fs::symlink_metadata(&dst).is_ok();
        if !present
            && !base.contains_key(&rel)
            && !own.has(KIND_FILE, &rel)
            && originals.is_empty()
            && matches!(intent, Ok(None))
        {
            continue;
        }
        let original_dir_link = symlink_parent(root, &root.join(ORIGINAL_DIR).join("probe"));
        let linked_original = originals.iter().any(|p| {
            symlink_parent(root, p)
                || fs::symlink_metadata(p).is_ok_and(|m| m.file_type().is_symlink())
        });
        let origin = own.get(KIND_FILE, &rel);
        let mut reason = None;
        let mut payload = None;
        if symlink_parent(root, &dst) || original_dir_link || linked_original {
            reason = Some("symlink parent or recovery link; provenance unknown");
        } else if fs::symlink_metadata(root.join(ownership::OWNERSHIP_FILE))
            .is_ok_and(|m| m.file_type().is_symlink())
            || origin_is_ambiguous(root, &rel)
        {
            reason = Some("malformed or conflicting ownership; provenance unknown");
        } else {
            for p in &originals {
                match fs::read(p) {
                    Ok(b) if payload.as_ref().is_none_or(|old| *old == b) => payload = Some(b),
                    _ => {
                        reason = Some("unreadable or conflicting originals; provenance unknown");
                        break;
                    }
                }
            }
        }
        let local = match fs::read(&dst) {
            Ok(b) => Some(b),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => None,
            Err(_) => {
                reason = Some("unreadable legacy file; provenance unknown");
                None
            }
        };
        let matches_base = local
            .as_ref()
            .zip(base.get(&rel))
            .is_some_and(|(b, h)| sha(b) == *h);
        let expected = payload
            .as_ref()
            .zip(base.get(&rel))
            .filter(|_| origin == ORIGIN_PREEXISTING)
            .map(|(original, base)| expected_intent(&rel, base, original));
        let valid_intent = intent
            .as_ref()
            .ok()
            .and_then(Option::as_ref)
            .zip(expected.as_ref())
            .is_some_and(|(saved, expected)| saved == expected);
        if reason.is_none() && (intent.is_err() || matches!(&intent, Ok(Some(_))) && !valid_intent)
        {
            reason = Some("invalid or stale restoration intent; provenance unknown");
        }
        let already_original = valid_intent
            && payload
                .as_ref()
                .zip(local.as_ref())
                .is_some_and(|(a, b)| a == b);
        let mut resolved = false;
        let mut action = None;
        if reason.is_none() {
            if origin == ORIGIN_CREATED && !originals.is_empty() {
                reason = Some("saved original contradicts created ownership; provenance unknown");
            } else if origin == ORIGIN_CREATED
                && originals.is_empty()
                && base.contains_key(&rel)
                && (matches_base || !present)
            {
                if present {
                    fs::remove_file(&dst).map_err(|e| format!("retire {rel}: {e}"))?;
                    action = Some("removed");
                }
                resolved = true;
            } else if let Some(original) = payload.as_ref().filter(|_| {
                origin == ORIGIN_PREEXISTING
                    && base.contains_key(&rel)
                    && (matches_base || already_original)
            }) {
                if !already_original {
                    if !valid_intent {
                        record_intent(root, &intent_path, expected.as_ref().unwrap())?;
                    }
                    write_file_atomic(&dst, original)?;
                    if fs::read(&dst).map_err(|e| format!("verify restored {rel}: {e}"))?
                        != *original
                    {
                        return Err(format!("verify restored {rel}: bytes differ"));
                    }
                }
                action = Some(if already_original {
                    "completed"
                } else {
                    "restored"
                });
                resolved = true;
            } else if base.contains_key(&rel) && local.is_some() && !matches_base {
                reason = Some("locally modified; retained with recovery");
            } else {
                reason = Some("provenance unknown or missing original; retained with recovery");
            }
        }
        // Persist only after a successful file disposition. Do not purge blob
        // evidence before this write succeeds; a failed restore never reaches it.
        let prior_base = base.clone();
        if base.contains_key(&rel) {
            let mut next = base.clone();
            next.remove(&rel);
            save_base_preserving_blobs(root, version, &next)?;
            *base = next;
        }
        if resolved {
            let prior_own = own.clone();
            let mut ledger_changed = false;
            let mut removed_originals = Vec::new();
            let cleanup = (|| {
                if own.has(KIND_FILE, &rel) {
                    let mut next = own.clone();
                    next.remove(KIND_FILE, &rel);
                    next.save(root)?;
                    *own = next;
                    ledger_changed = true;
                }
                for p in &originals {
                    remove_backup(p)?;
                    removed_originals.push(p.clone());
                }
                // Last: a cleanup failure before this point keeps valid intent
                // and rollback restores its matching base/ledger/originals.
                if valid_intent || expected.is_some() && action == Some("restored") {
                    #[cfg(test)]
                    fail_intent_operation(&intent_path, true)?;
                    fs::remove_file(&intent_path)
                        .map_err(|e| format!("remove retirement intent: {e}"))?;
                }
                Ok::<(), String>(())
            })();
            if let Err(error) = cleanup {
                let failures = rollback_cleanup(
                    root,
                    version,
                    &prior_base,
                    ledger_changed.then_some(&prior_own),
                    &removed_originals,
                    payload.as_ref(),
                );
                *base = prior_base;
                *own = prior_own;
                if failures.is_empty() {
                    out.push_str(&format!("notice    {rel}: file disposition finished; prior recovery evidence preserved; retry cleanup\n"));
                    return Err(error);
                }
                out.push_str(&format!("notice    {rel}: cleanup and recovery incomplete; retain snapshots and inspect metadata before retry\n"));
                return Err(format!(
                    "{error}; recovery also failed: {}",
                    failures.join("; ")
                ));
            }
            if let Some(action) = action {
                out.push_str(&format!("{action:<10}{rel} (legacy procedure retired)\n"));
            }
        } else {
            out.push_str(&format!("KEPT      {rel} ({})\n", reason.unwrap()));
        }
    }
    let intents = root.join(INTENT_DIR);
    if !symlink_parent(root, &intents.join("probe")) {
        super::remove_dir_if_empty(&intents);
    }
    let dir = root.join(PLAYBOOK_DIR_TGT);
    if !symlink_parent(root, &dir.join("probe")) {
        super::remove_dir_if_empty(&dir);
    }
    Ok(())
}
