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

#[cfg(test)]
mod tests {
    use std::fs;
    use std::path::Path;

    use super::super::registry::{canonical_root, registered};
    use super::super::{all_targets, PLAYBOOK_DIR_TGT, WORKFLOW_TARGET};
    use super::*;
    use crate::test_support::{
        install_ok, read_file, set_registry, temp_repo, tree_snapshot, write_file, IsolatedEnv,
    };

    #[test]
    fn check_reports_each_drift_kind() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);

        let stale = all_targets()
            .into_iter()
            .find(|t| t.dst.starts_with(PLAYBOOK_DIR_TGT))
            .expect("the managed set has playbooks")
            .dst;
        write_file(root, &stale, "old playbook\n");
        fs::remove_file(root.join(WORKFLOW_TARGET)).unwrap();
        write_file(
            root,
            AGENTS_TARGET,
            "<!-- ZHARNESS:BEGIN -->\nold block\n<!-- ZHARNESS:END -->\n",
        );
        write_file(
            root,
            PROJECT_TARGET,
            "# Project\n\n## What is this project?\nx\n",
        );

        let got = check(root).unwrap();
        let joined = got.join("\n");
        for want in [
            format!("stale    {stale}"),
            format!("missing  {WORKFLOW_TARGET}"),
            format!("stale    {AGENTS_TARGET} block"),
            format!("heading  {PROJECT_TARGET} lacks \"## Who is it for?\""),
        ] {
            assert!(joined.contains(&want), "missing {want:?} in:\n{joined}");
        }
        assert!(
            !joined.contains("lacks \"## What is this project?\""),
            "present heading reported as missing:\n{joined}"
        );
    }

    #[test]
    fn check_is_read_only() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        write_file(root, WORKFLOW_TARGET, "edited\n");
        let before = tree_snapshot(root);

        let mut out = String::new();
        let _ = run_check(root, false, &mut out);
        assert_eq!(tree_snapshot(root), before, "check modified the repository");
    }

    #[test]
    fn run_check_all_reports_missing_and_fails_on_drift() {
        let _env = IsolatedEnv::new();
        let clean = temp_repo();
        let drifted = temp_repo();
        install_ok(clean.path());
        install_ok(drifted.path());
        write_file(drifted.path(), WORKFLOW_TARGET, "edited\n");
        let gone = temp_repo().path().join("deleted-repo");

        let mut roots = registered().unwrap();
        roots.push(gone.to_string_lossy().into_owned());
        set_registry(&roots).unwrap();

        let mut out = String::new();
        let err = run_check(Path::new(""), true, &mut out);
        assert_eq!(
            err,
            Err(ERR_DRIFT.to_string()),
            "run_check --all = {err:?}, want ErrDrift\n{out}"
        );
        for want in [
            format!("current  {}", canonical_root(clean.path()).display()),
            format!("drift    {}", canonical_root(drifted.path()).display()),
            format!("missing: {}", gone.display()),
            "1 repository(ies) drifted".to_string(),
        ] {
            assert!(out.contains(&want), "missing {want:?} in:\n{out}");
        }
    }

    #[test]
    fn run_check_all_missing_only_is_clean() {
        let _env = IsolatedEnv::new();
        let gone = temp_repo().path().join("gone");
        set_registry(&[gone.to_string_lossy().into_owned()]).unwrap();

        let mut out = String::new();
        run_check(Path::new(""), true, &mut out)
            .unwrap_or_else(|e| panic!("missing roots alone must not fail: {e}\n{out}"));
    }

    #[test]
    fn run_check_all_continues_past_unreadable_root() {
        let _env = IsolatedEnv::new();
        let broken = temp_repo();
        let good = temp_repo();
        install_ok(broken.path());
        install_ok(good.path());
        let wf = broken.path().join(WORKFLOW_TARGET);
        fs::remove_file(&wf).unwrap();
        fs::create_dir(&wf).unwrap();

        let mut out = String::new();
        let err = run_check(Path::new(""), true, &mut out);
        assert!(
            err.is_err() && err != Err(ERR_DRIFT.to_string()),
            "run_check = {err:?}, want a could-not-check error\n{out}"
        );
        for want in [
            format!("error    {}", canonical_root(broken.path()).display()),
            format!("current  {}", canonical_root(good.path()).display()),
        ] {
            assert!(
                out.contains(&want),
                "expected {want:?} (error line and the next root checked):\n{out}"
            );
        }
    }

    #[test]
    fn check_reports_a_clean_install_as_current() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);

        let got = check(root).unwrap();
        assert!(got.is_empty(), "drift on a fresh install: {got:?}");
        let mut out = String::new();
        run_check(root, false, &mut out).unwrap_or_else(|e| panic!("run_check clean = {e}\n{out}"));
        assert!(!read_file(root, WORKFLOW_TARGET).is_empty());
    }
}
