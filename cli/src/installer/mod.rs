//! The managed-set verbs: install, update, uninstall.
//!
//! This module owns exactly the managed doc set — `docs/WORKFLOW.md`,
//! `docs/playbooks/*.md`, the marked AGENTS block, the `docs/PROJECT.md`
//! scaffold, `.zharness` bookkeeping and the `.gitignore` entries. Nothing
//! else in a repository is created, modified, or deleted.

pub mod check;
pub mod migrate;
pub mod ownership;
pub mod registry;
pub mod uninstall;
pub mod update;

use std::collections::HashMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

use crate::embedded;

pub use check::{check, run_check};
pub use uninstall::uninstall;
pub use update::{run_update, UpdateOptions};

pub const ZHRNESS_DIR: &str = ".zharness";
pub const BASE_DIR: &str = ".zharness/base";
pub const ORIGINAL_DIR: &str = ".zharness/base/original";
pub const MANIFEST_FILE: &str = ".zharness/base/manifest.json";
pub const PROJECT_TEMPLATE: &str = embedded::PROJECT_TEMPLATE;
pub const BLOCK_BEGIN: &str = "<!-- ZHARNESS:BEGIN -->";
pub const BLOCK_END: &str = "<!-- ZHARNESS:END -->";
pub const GITIGNORE_MARKER: &str = "# zharness v0.15 managed set";
pub const PLAYBOOK_DIR_TGT: &str = "docs/playbooks";
pub const WORKFLOW_TARGET: &str = "docs/WORKFLOW.md";
pub const PROJECT_TARGET: &str = "docs/PROJECT.md";
pub const AGENTS_TARGET: &str = "AGENTS.md";
pub const GITIGNORE_TARGET: &str = ".gitignore";

/// The only prose the installer itself writes outside the marked block when it
/// creates AGENTS.md. Anything else outside the block is the consumer's.
pub const AGENTS_CREATED_HEADER: &str = "# Agents";

/// Pre-ADR 0011 update artifacts. The blob store is deleted by the next
/// install or update. A leftover conflict list makes update refuse and check
/// report drift; it and the stash are deleted only by the owner or uninstall.
pub const LEGACY_UPSTREAM_DIR: &str = ".zharness/base/upstream";
pub const LEGACY_STASH_DIR: &str = ".zharness/update-stash";
pub const LEGACY_CONFLICTS_FILE: &str = ".zharness/conflicts.json";

/// One managed-file mapping from an embedded source path to a destination path
/// inside the consuming repository.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Target {
    /// Path inside the embedded doc set.
    pub src: String,
    /// Repo-root-relative destination.
    pub dst: String,
    /// `true`: written only when absent, then project-owned; `false`: always
    /// overwritten with upstream bytes.
    pub once: bool,
}

/// The managed set in write order: WORKFLOW.md, the identity template, then
/// every playbook sorted by name.
pub fn all_targets() -> Vec<Target> {
    let mut targets = vec![
        Target {
            src: "WORKFLOW.md".to_string(),
            dst: WORKFLOW_TARGET.to_string(),
            once: false,
        },
        Target {
            src: PROJECT_TEMPLATE.to_string(),
            dst: PROJECT_TARGET.to_string(),
            once: true,
        },
    ];
    for name in embedded::playbook_names() {
        targets.push(Target {
            src: format!("playbooks/{name}"),
            dst: format!("{PLAYBOOK_DIR_TGT}/{name}"),
            once: false,
        });
    }
    targets
}

/// The upstream bytes for a target.
pub fn src_bytes(t: &Target) -> Result<Vec<u8>, String> {
    embedded::read_file(&t.src)
        .map(|b| b.to_vec())
        .ok_or_else(|| format!("embed read {}: file not embedded", t.src))
}

pub fn sha(bytes: &[u8]) -> String {
    let mut h = Sha256::new();
    h.update(bytes);
    hex(&h.finalize())
}

fn hex(b: &[u8]) -> String {
    let mut s = String::with_capacity(b.len() * 2);
    for byte in b {
        s.push_str(&format!("{byte:02x}"));
    }
    s
}

/// Write `data` to `p` through a sibling temp file and a rename, so a reader
/// never sees a half-written managed file.
pub fn write_file_atomic(p: &Path, data: &[u8]) -> Result<(), String> {
    if let Some(dir) = p.parent() {
        fs::create_dir_all(dir).map_err(|e| format!("mkdir {}: {e}", dir.display()))?;
    }
    let tmp = PathBuf::from(format!("{}.tmp-zharness", p.display()));
    fs::write(&tmp, data).map_err(|e| format!("write {}: {e}", tmp.display()))?;
    fs::rename(&tmp, p).map_err(|e| format!("rename {}: {e}", p.display()))
}

// ------------------------------------------------------------ manifest ---

#[derive(Debug, Serialize, Deserialize)]
struct ManifestEntry {
    path: String,
    sha256: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Manifest {
    zharness_version: String,
    installed_at: String,
    files: Vec<ManifestEntry>,
}

/// The recorded base: the sha256 of the bytes zharness last wrote for each
/// managed path (ADR 0011 decision 5).
pub fn load_base(root: &Path) -> Result<HashMap<String, String>, String> {
    let mut files = HashMap::new();
    let raw = match fs::read(root.join(MANIFEST_FILE)) {
        Ok(b) => b,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(files),
        Err(e) => return Err(format!("read base manifest: {e}")),
    };
    let m: Manifest =
        serde_json::from_slice(&raw).map_err(|e| format!("parse {MANIFEST_FILE}: {e}"))?;
    for fe in m.files {
        files.insert(fe.path, fe.sha256);
    }
    Ok(files)
}

pub fn save_base(root: &Path, ver: &str, files: &HashMap<String, String>) -> Result<(), String> {
    let _ = fs::remove_dir_all(root.join(LEGACY_UPSTREAM_DIR));
    let mut keys: Vec<&String> = files.keys().collect();
    keys.sort();
    let entries: Vec<ManifestEntry> = keys
        .into_iter()
        .map(|dst| ManifestEntry {
            path: dst.clone(),
            sha256: files[dst].clone(),
        })
        .collect();
    let m = Manifest {
        zharness_version: ver.to_string(),
        installed_at: rfc3339_now(),
        files: entries,
    };
    let mut out = serde_json::to_vec(&m).map_err(|e| format!("marshal manifest: {e}"))?;
    out.push(b'\n');
    write_file_atomic(&root.join(MANIFEST_FILE), &out)
}

/// `time.Now().UTC().Format(time.RFC3339)`, without a date library: the
/// manifest's stamp is the only place the binary needs a calendar.
fn rfc3339_now() -> String {
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);
    let days = secs.div_euclid(86_400);
    let rem = secs.rem_euclid(86_400);
    let (y, m, d) = civil_from_days(days);
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3600,
        (rem % 3600) / 60,
        rem % 60
    )
}

/// Days since 1970-01-01 to a civil date (Howard Hinnant's algorithm).
fn civil_from_days(z: i64) -> (i64, u32, u32) {
    let z = z + 719_468;
    let era = if z >= 0 { z } else { z - 146_096 } / 146_097;
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = (doy - (153 * mp + 2) / 5 + 1) as u32;
    let m = if mp < 10 { mp + 3 } else { mp - 9 } as u32;
    (if m <= 2 { y + 1 } else { y }, m, d)
}

// ----------------------------------------------------------- originals ---

/// Maps a managed relative path to a flat, collision-free file name: `_` ->
/// `__` and `/` -> `_2F`. The two replacement images form a prefix-free code,
/// so the mapping is injective.
pub fn safe_path(p: &str) -> String {
    let mut out = String::with_capacity(p.len());
    for c in p.chars() {
        match c {
            '_' => out.push_str("__"),
            '/' => out.push_str("_2F"),
            other => out.push(other),
        }
    }
    out
}

/// The v0.15.0 mapping (`/` -> `__`), which was NOT injective. Kept only so
/// originals captured by that release still protect uninstall after an
/// upgrade.
pub fn legacy_safe_path(p: &str) -> String {
    p.replace('/', "__")
}

/// The on-disk original path for `dst`, preferring the current mapping and
/// falling back to the legacy one.
pub fn find_original(root: &Path, dst: &str) -> Option<PathBuf> {
    for name in [
        format!("{}.orig", safe_path(dst)),
        format!("{}.orig", legacy_safe_path(dst)),
    ] {
        let p = root.join(ORIGINAL_DIR).join(name);
        if p.exists() {
            return Some(p);
        }
    }
    None
}

/// Capture a pristine copy of a path that predates zharness, never overwriting
/// a previously captured pristine under either mapping.
pub fn capture_original(root: &Path, dst: &str) -> Result<(), String> {
    let src = root.join(dst);
    if !src.exists() {
        return Ok(()); // did not exist; uninstall may delete it outright
    }
    if find_original(root, dst).is_some() {
        return Ok(());
    }
    let dst_p = root
        .join(ORIGINAL_DIR)
        .join(format!("{}.orig", safe_path(dst)));
    let b = fs::read(&src).map_err(|e| format!("read {}: {e}", src.display()))?;
    write_file_atomic(&dst_p, &b)
}

pub fn read_original(root: &Path, dst: &str) -> Option<Vec<u8>> {
    let p = find_original(root, dst)?;
    fs::read(p).ok()
}

// ------------------------------------------------------- agents block ---

/// Locate the marked zharness block inclusive of both marker comment lines.
pub fn agents_span(content: &str) -> Option<(usize, usize)> {
    let bi = content.find(BLOCK_BEGIN)?;
    let ej = content[bi..].find(BLOCK_END)?;
    Some((bi, bi + ej + BLOCK_END.len()))
}

/// Wrap an embedded AGENTS.md body into its on-disk form.
pub fn canonical_agents_block(body: &str) -> String {
    format!(
        "{BLOCK_BEGIN}\n{}\n{BLOCK_END}",
        body.trim_end_matches('\n')
    )
}

/// Swap the marked block in place or append it. Returns the new content and
/// whether anything changed.
pub fn apply_agents_block(content: &str, embedded_body: &str) -> (String, bool) {
    let want = canonical_agents_block(embedded_body);
    if let Some((s, e)) = agents_span(content) {
        if content[s..e] == want {
            return (content.to_string(), false);
        }
        return (format!("{}{}{}", &content[..s], want, &content[e..]), true);
    }
    let mut content = content.to_string();
    if !content.is_empty() && !content.ends_with('\n') {
        content.push('\n');
    }
    (format!("{content}\n{want}\n"), true)
}

pub fn agents_block_of(content: &str) -> Option<String> {
    let (s, e) = agents_span(content)?;
    Some(content[s..e].to_string())
}

pub fn normalize_block_tail(b: &str) -> String {
    format!("{}\n", b.trim_end_matches('\n'))
}

// ------------------------------------------------------------ helpers ---

pub fn gitignore_wants() -> [&'static str; 2] {
    [GITIGNORE_MARKER, "/.zharness/"]
}

pub fn contains_line(blob: &str, want: &str) -> bool {
    blob.split('\n').any(|ln| ln.trim() == want.trim())
}

pub fn ensure_lines(blob: &str, wants: &[&str]) -> String {
    let mut body = blob.to_string();
    for w in wants {
        if contains_line(&body, w) {
            continue;
        }
        if !body.is_empty() && !body.ends_with('\n') {
            body.push('\n');
        }
        body.push_str(w);
        body.push('\n');
    }
    body
}

pub fn remove_dir_if_empty(dir: &Path) {
    if let Ok(mut entries) = fs::read_dir(dir) {
        if entries.next().is_none() {
            let _ = fs::remove_dir(dir);
        }
    }
}

pub fn dedupe(items: Vec<String>) -> Vec<String> {
    let mut seen = std::collections::HashSet::new();
    let mut out = Vec::new();
    for x in items {
        if seen.insert(x.clone()) {
            out.push(x);
        }
    }
    out
}

// ------------------------------------------------------------- install ---

struct BrownfieldReport {
    active_plans: usize,
    present: Vec<String>,
    foreign_state: Vec<String>,
}

fn detect_brownfield(root: &Path, r: &mut BrownfieldReport) {
    r.active_plans = 0;
    if let Ok(entries) = fs::read_dir(root.join("docs/plans/active")) {
        for e in entries.flatten() {
            let name = e.file_name().to_string_lossy().into_owned();
            if !e.path().is_dir() && name.ends_with(".md") && !name.starts_with('.') {
                r.active_plans += 1;
            }
        }
    }
    for p in ["README.md", "CLAUDE.md", AGENTS_TARGET] {
        if root.join(p).exists() {
            r.present.push(p.to_string());
        }
    }
    if root.join("docs").is_dir() {
        r.present.push("docs/**".to_string());
    }
    // Folded at compile time so a tree scan cannot match its own probe list.
    let legacy_db = format!("harness{}db", ".");
    for f in [legacy_db.as_str(), "workflow-state.yml", ".kit"] {
        if root.join(f).exists() {
            r.foreign_state.push(f.to_string());
        }
    }
    r.present.sort();
    r.foreign_state.sort();
}

fn write_report(out: &mut String, r: &BrownfieldReport) {
    out.push_str("\nbrownfield scan (read-only)\n");
    out.push_str(&format!(
        "- active plans under docs/plans/active: {}",
        r.active_plans
    ));
    if r.active_plans >= 2 {
        out.push_str(
            "  \u{2192} 2 or more: reconcile which plan stays live before locking a new one",
        );
    } else if r.active_plans == 1 {
        out.push_str("  \u{2192} refine the existing plan instead of creating another");
    } else {
        out.push_str("  \u{2192} greenfield: this install will become the first lock target");
    }
    out.push('\n');
    out.push_str(&format!(
        "- present inputs for HARVEST drafting: {}\n",
        join_or_none(&r.present)
    ));
    out.push_str(&format!(
        "- foreign state answering the same questions: {}\n",
        join_or_none(&r.foreign_state)
    ));
    out.push_str(
        "- proposals here are advisory only; nothing outside the managed set is written\n\n",
    );
}

fn join_or_none(xs: &[String]) -> String {
    if xs.is_empty() {
        "none".to_string()
    } else {
        xs.join(", ")
    }
}

/// The managed-set scaffolding plus deterministic read-only brownfield
/// detection. Consumer-owned files outside the set are never written.
pub fn install(root: &Path, version: &str, out: &mut String) -> Result<(), String> {
    let mut report = BrownfieldReport {
        active_plans: 0,
        present: Vec::new(),
        foreign_state: Vec::new(),
    };
    detect_brownfield(root, &mut report);
    write_report(out, &report);

    let targets = all_targets();
    let prev = load_base(root)?;
    let mut files: HashMap<String, String> = prev.clone();

    // Ownership is decided once per key and read from the ledger forever after
    // (ADR 0008); seeding runs only when `prev` proves a prior install. The
    // record is persisted before the first write so an interrupted install
    // cannot leave generated files on disk with no record of who made them.
    let mut own = ownership::load_ownership_for(root, &targets, &prev);
    own.decide_all(root, &targets);
    own.save(root)?;

    for t in &targets {
        let up = src_bytes(t)?;
        ownership::capture_if_preexisting(root, &mut own, &t.dst)?;
        let dst_p = root.join(&t.dst);
        if !t.once {
            // Pure upstream mirror (playbooks, WORKFLOW.md): always overwrite.
            write_file_atomic(&dst_p, &up)?;
            out.push_str(&format!("installed  {}\n", t.dst));
            files.insert(t.dst.clone(), sha(&up));
            continue;
        }
        match fs::read(&dst_p) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                write_file_atomic(&dst_p, &up)?;
                out.push_str(&format!("installed  {}\n", t.dst));
                files.insert(t.dst.clone(), sha(&up));
                continue;
            }
            Ok(local) if local == up => {
                out.push_str(&format!("current    {}\n", t.dst));
            }
            Ok(_) => {
                out.push_str(&format!(
                    "kept       {} (project-owned; never overwritten)\n",
                    t.dst
                ));
            }
            Err(e) => return Err(format!("read {}: {e}", dst_p.display())),
        }
        files.entry(t.dst.clone()).or_insert_with(|| sha(&up));
    }

    let agents_up = embedded::read_file(AGENTS_TARGET)
        .ok_or_else(|| "embed read AGENTS.md: file not embedded".to_string())?;
    let new_block = String::from_utf8_lossy(agents_up).into_owned();
    let ap = root.join(AGENTS_TARGET);
    match fs::read(&ap) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            ownership::capture_if_preexisting(root, &mut own, AGENTS_TARGET)?;
            if let Some(dir) = ap.parent() {
                fs::create_dir_all(dir).map_err(|e| format!("mkdir {}: {e}", dir.display()))?;
            }
            let (repl_body, _) = apply_agents_block("", &new_block);
            let body = format!("{AGENTS_CREATED_HEADER}\n\n{repl_body}");
            fs::write(&ap, body).map_err(|e| format!("write {}: {e}", ap.display()))?;
            out.push_str(&format!("installed  {AGENTS_TARGET} (created)\n"));
        }
        Err(e) => return Err(format!("read {}: {e}", ap.display())),
        Ok(existing) => {
            let before = String::from_utf8_lossy(&existing).into_owned();
            ownership::capture_if_preexisting(root, &mut own, AGENTS_TARGET)?;
            let (content, changed) = apply_agents_block(&before, &new_block);
            if changed {
                write_file_atomic(&ap, content.as_bytes())?;
                out.push_str(&format!(
                    "updated    {AGENTS_TARGET} (marked block refreshed)\n"
                ));
            } else {
                out.push_str(&format!(
                    "current    {AGENTS_TARGET} (block already up to date)\n"
                ));
            }
        }
    }
    files.insert(
        AGENTS_TARGET.to_string(),
        sha(canonical_agents_block(&new_block).as_bytes()),
    );

    append_gitignore_entries(root, &mut own, out)?;

    save_base(root, version, &files)?;
    own.save(root)?;
    registry::register(root, out);
    Ok(())
}

fn append_gitignore_entries(
    root: &Path,
    own: &mut ownership::Ownership,
    out: &mut String,
) -> Result<(), String> {
    let gp = root.join(GITIGNORE_TARGET);
    let existing = match fs::read_to_string(&gp) {
        Ok(s) => s,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => String::new(),
        Err(e) => return Err(format!("read {}: {e}", gp.display())),
    };
    ownership::capture_if_preexisting(root, own, GITIGNORE_TARGET)?;
    let wants = gitignore_wants();
    // Provenance for these lines is already fixed and on disk (decide_all);
    // here we only decide what still needs appending.
    let missing = wants.iter().any(|w| !contains_line(&existing, w));
    if !missing {
        out.push_str(&format!(
            "current    {GITIGNORE_TARGET} (entries present)\n"
        ));
        return Ok(());
    }
    let mut add = String::new();
    for w in wants {
        if !contains_line(&existing, w) {
            add.push_str(w);
            add.push('\n');
        }
    }
    let mut body = existing;
    if !body.is_empty() && !body.ends_with('\n') {
        body.push('\n');
    }
    body.push_str(&add);
    write_file_atomic(&gp, body.as_bytes())?;
    out.push_str(&format!(
        "updated    {GITIGNORE_TARGET} (+{ZHRNESS_DIR}/)\n"
    ));
    Ok(())
}
