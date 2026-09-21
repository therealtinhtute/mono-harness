//! Drift detection: compares a repository's managed set with the embedded
//! bytes without writing anything.

use std::fs;
use std::path::Path;

use super::{
    agents_block_of, all_targets, canonical_agents_block, normalize_block_tail, registry,
    src_bytes, Target, AGENTS_TARGET, LEGACY_CONFLICTS_FILE, PROJECT_TARGET, PROJECT_TEMPLATE,
};

/// The message the check verbs return when any repository lags the binary.
pub const ERR_DRIFT: &str = "managed docs drift from this zharness binary";

/// One line per drifted item, empty when the repository is current.
pub fn check(root: &Path) -> Result<Vec<String>, String> {
    let targets = all_targets();
    let mut drift = Vec::new();
    if root.join(LEGACY_CONFLICTS_FILE).exists() {
        drift.push(format!(
            "conflict {LEGACY_CONFLICTS_FILE} (unresolved pre-0011 update)"
        ));
    }
    for t in &targets {
        if t.once {
            continue;
        }
        let want = src_bytes(t)?;
        match fs::read(root.join(&t.dst)) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                drift.push(format!("missing  {}", t.dst));
            }
            Err(e) => return Err(format!("read {}: {e}", t.dst)),
            Ok(got) if got != want => drift.push(format!("stale    {}", t.dst)),
            Ok(_) => {}
        }
    }

    let body = src_bytes(&Target {
        src: AGENTS_TARGET.to_string(),
        dst: AGENTS_TARGET.to_string(),
        once: false,
    })?;
    let agents = fs::read(root.join(AGENTS_TARGET)).unwrap_or_default();
    match agents_block_of(&String::from_utf8_lossy(&agents)) {
        None => drift.push(format!("missing  {AGENTS_TARGET} block")),
        Some(block) => {
            let want = canonical_agents_block(&String::from_utf8_lossy(&body));
            if normalize_block_tail(&block) != normalize_block_tail(&want) {
                drift.push(format!("stale    {AGENTS_TARGET} block"));
            }
        }
    }

    let tmpl = src_bytes(&Target {
        src: PROJECT_TEMPLATE.to_string(),
        dst: PROJECT_TARGET.to_string(),
        once: true,
    })?;
    match fs::read(root.join(PROJECT_TARGET)) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            drift.push(format!("missing  {PROJECT_TARGET}"));
        }
        Err(e) => return Err(format!("read {PROJECT_TARGET}: {e}")),
        Ok(project) => {
            let have = headings(&project);
            for h in headings(&tmpl) {
                if !have.contains(&h) {
                    drift.push(format!("heading  {PROJECT_TARGET} lacks {h:?}"));
                }
            }
        }
    }
    Ok(drift)
}

fn headings(b: &[u8]) -> Vec<String> {
    String::from_utf8_lossy(b)
        .lines()
        .map(|l| l.trim_end_matches([' ', '\t']).to_string())
        .filter(|l| l.starts_with("## "))
        .collect()
}

/// Reports drift for one root, or for every registered root when `all` is set.
/// Registered roots that no longer exist are reported, not failed.
pub fn run_check(root: &Path, all: bool, out: &mut String) -> Result<(), String> {
    let roots: Vec<String> = if all {
        let reg = registry::registered()?;
        if reg.is_empty() {
            out.push_str("no registered repositories\n");
            return Ok(());
        }
        reg
    } else {
        vec![root.to_string_lossy().into_owned()]
    };

    let mut drifted = 0usize;
    let mut failed = 0usize;
    for r in &roots {
        let p = Path::new(r);
        if !p.is_dir() {
            out.push_str(&format!("missing: {r}\n"));
            continue;
        }
        match check(p) {
            Err(e) => {
                failed += 1;
                out.push_str(&format!("error    {r}: {e}\n"));
            }
            Ok(lines) if lines.is_empty() => out.push_str(&format!("current  {r}\n")),
            Ok(lines) => {
                drifted += 1;
                out.push_str(&format!("drift    {r}\n"));
                for l in lines {
                    out.push_str(&format!("  {l}\n"));
                }
            }
        }
    }
    if failed > 0 {
        return Err(format!("{failed} repository(ies) could not be checked"));
    }
    if drifted > 0 {
        out.push_str(&format!(
            "{drifted} repository(ies) drifted — run `zharness update` in each.\n"
        ));
        return Err(ERR_DRIFT.to_string());
    }
    Ok(())
}
