//! The 9-section to 5-section plan migration.
//!
//! A plan is migrated only when its `## ` headings are exactly the legacy set.
//! The `## Validation` body is copied verbatim and checked afterwards, so a
//! migration can never alter proof evidence.

use std::collections::HashMap;
use std::fs;
use std::path::Path;

use super::write_file_atomic;

pub const ACTIVE_PLANS_DIR: &str = "docs/plans/active";

/// The 9-section plan format that predates the 5-section one.
const LEGACY_PLAN_SECTIONS: [&str; 9] = [
    "Outcome",
    "Authority and Requirements",
    "Non-goals",
    "Approach and Risks",
    "Phases and Verification",
    "Progress",
    "Decisions",
    "Validation",
    "Current State and Next Action",
];

const SLIM_PLAN_SECTIONS: [&str; 5] = [
    "Goal",
    "Phases and Verification",
    "Log",
    "Validation",
    "Current State and Next Action",
];

/// Rewrite a 9-section plan into the 5-section format by merging sections.
/// Returns the new bytes and whether anything changed.
pub fn migrate_plan(input: &[u8]) -> Result<(Vec<u8>, bool), String> {
    let s = String::from_utf8_lossy(input).into_owned();
    let (front, preamble, sections) = match split_plan(&s) {
        Some(t) => t,
        None => return Ok((input.to_vec(), false)),
    };
    if sections.len() != LEGACY_PLAN_SECTIONS.len() {
        return Ok((input.to_vec(), false));
    }
    for name in LEGACY_PLAN_SECTIONS {
        if !sections.contains_key(name) {
            return Ok((input.to_vec(), false));
        }
    }

    let mut fm = String::new();
    for l in split_after(&front, '\n') {
        if l.starts_with("intake_id:") || l.starts_with("story_id:") {
            continue;
        }
        fm.push_str(l);
    }

    let mut phases = String::new();
    for l in split_after(&sections["Phases and Verification"], '\n') {
        let body = l.strip_suffix('\n').unwrap_or(l);
        let lead = body.len() - body.trim_start_matches([' ', '\t']).len();
        if body[lead..].starts_with("- story_id:") {
            continue;
        }
        phases.push_str(&strip_status_bullet(l));
    }

    let mut out = String::new();
    out.push_str(&fm);
    out.push_str(&preamble);
    out.push_str("## Goal\n");
    out.push_str(&sections["Outcome"]);
    out.push_str("### Authority and Requirements\n");
    out.push_str(&sections["Authority and Requirements"]);
    out.push_str("### Non-goals\n");
    out.push_str(&sections["Non-goals"]);
    out.push_str("## Phases and Verification\n");
    out.push_str(&sections["Approach and Risks"]);
    out.push_str(&phases);
    out.push_str("## Log\n");
    out.push_str(&sections["Progress"]);
    out.push_str("### Decisions\n");
    out.push_str(&sections["Decisions"]);
    out.push_str("## Validation\n");
    out.push_str(&sections["Validation"]);
    out.push_str("## Current State and Next Action\n");
    out.push_str(&sections["Current State and Next Action"]);

    if let Some((_, _, after)) = split_plan(&out) {
        if after.get("Validation") != sections.get("Validation") {
            return Err("plan migration would change ## Validation".to_string());
        }
    }
    Ok((out.into_bytes(), true))
}

/// Turn each phase's `- status:` bullet into the bare `status:` line the
/// completed-plan guard reads.
fn strip_status_bullet(line: &str) -> String {
    let (body, nl) = match line.strip_suffix('\n') {
        Some(b) => (b, "\n"),
        None => (line, ""),
    };
    let lead = body.len() - body.trim_start_matches([' ', '\t']).len();
    let (indent, rest) = body.split_at(lead);
    match rest.strip_prefix("- status:") {
        Some(status) => format!("{indent}status:{status}{nl}"),
        None => format!("{body}{nl}"),
    }
}

/// `strings.SplitAfter`: every piece keeps its trailing newline, and a string
/// ending in a newline produces no trailing empty piece.
fn split_after(s: &str, sep: char) -> Vec<&str> {
    let mut out = Vec::new();
    let mut start = 0;
    for (i, c) in s.char_indices() {
        if c == sep {
            out.push(&s[start..i + c.len_utf8()]);
            start = i + c.len_utf8();
        }
    }
    if start < s.len() {
        out.push(&s[start..]);
    }
    out
}

/// Separate optional frontmatter, the text before the first `## ` heading, and
/// each `## ` section body keyed by heading text. Every body ends with a
/// newline so sections can be concatenated in any order. A repeated heading
/// returns `None`.
pub fn split_plan(s: &str) -> Option<(String, String, HashMap<String, String>)> {
    let mut s = s;
    let mut front = String::new();
    if let Some(rest) = s.strip_prefix("---\n") {
        if let Some(end) = rest.find("\n---\n") {
            let cut = 4 + end + 5;
            front = s[..cut].to_string();
            s = &s[cut..];
        }
    }

    let mut sections: HashMap<String, String> = HashMap::new();
    let mut preamble = String::new();
    let mut cur: Option<String> = None;
    let mut body = String::new();

    for l in split_after(s, '\n') {
        let trimmed = l.strip_suffix('\n').unwrap_or(l);
        if let Some(h) = trimmed.strip_prefix("## ") {
            flush(&mut sections, &mut preamble, &mut cur, &mut body)?;
            cur = Some(h.to_string());
            continue;
        }
        body.push_str(l);
    }
    flush(&mut sections, &mut preamble, &mut cur, &mut body)?;
    Some((front, preamble, sections))
}

fn flush(
    sections: &mut HashMap<String, String>,
    preamble: &mut String,
    cur: &mut Option<String>,
    body: &mut String,
) -> Option<()> {
    match cur {
        None => *preamble = std::mem::take(body),
        Some(name) => {
            if sections.contains_key(name) {
                return None;
            }
            let mut t = std::mem::take(body);
            if !t.ends_with('\n') {
                t.push('\n');
            }
            sections.insert(name.clone(), t);
        }
    }
    Some(())
}

fn is_slim_plan(secs: &HashMap<String, String>) -> bool {
    SLIM_PLAN_SECTIONS.iter().all(|h| secs.contains_key(*h)) && secs.len() == 5
}

/// The pending write, if any, for the one active plan.
#[derive(Debug, Default)]
pub struct PlanMigration {
    pub path: String,
    pub data: Vec<u8>,
    pub notice: String,
}

/// Decide what update does to the active plan before anything is written:
/// migrate it, leave it with a notice, or refuse.
pub fn prepare_plan_migration(root: &Path) -> Result<PlanMigration, String> {
    let entries = match fs::read_dir(root.join(ACTIVE_PLANS_DIR)) {
        Ok(e) => e,
        Err(_) => return Ok(PlanMigration::default()),
    };
    let mut plans: Vec<String> = Vec::new();
    for e in entries.flatten() {
        let name = e.file_name().to_string_lossy().into_owned();
        if e.path().is_dir() || !name.ends_with(".md") || name.starts_with('.') {
            continue;
        }
        plans.push(format!("{ACTIVE_PLANS_DIR}/{name}"));
    }
    plans.sort();
    match plans.len() {
        0 => return Ok(PlanMigration::default()),
        1 => {}
        n => {
            return Ok(PlanMigration {
                notice: format!("{n} active plans; none migrated"),
                ..Default::default()
            })
        }
    }

    let input = fs::read(root.join(&plans[0])).map_err(|e| format!("read {}: {e}", plans[0]))?;
    let (out, changed) = migrate_plan(&input).map_err(|e| format!("{}: {e}", plans[0]))?;
    if changed {
        return Ok(PlanMigration {
            path: plans[0].clone(),
            data: out,
            notice: String::new(),
        });
    }
    match split_plan(&String::from_utf8_lossy(&input)) {
        Some((_, _, secs)) if is_slim_plan(&secs) => Ok(PlanMigration::default()),
        _ => Ok(PlanMigration {
            notice: format!("{} has an unrecognized section set; not migrated", plans[0]),
            ..Default::default()
        }),
    }
}

/// Write a pending migration. Split out so `run_update` can decide everything
/// before it writes anything.
pub fn apply_plan_migration(root: &Path, mig: &PlanMigration) -> Result<(), String> {
    if mig.path.is_empty() {
        return Ok(());
    }
    write_file_atomic(&root.join(&mig.path), &mig.data)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::installer::sha;
    use crate::test_support::{
        install_ok, read_file, temp_repo, update_ok, write_file, IsolatedEnv,
    };

    /// The legacy `## Validation` body, byte for byte as a pre-5-section plan
    /// carried it. Migration must copy it verbatim.
    const LEGACY_VALIDATION: &str = "- 2026-09-19T08:29Z — phase `p1` — verdict: APPROVED — mode: gate\n  - `true` — ok\n  - judge: independent\n";

    fn legacy_body(h: &str) -> String {
        match h {
            "Outcome" => "- result: x\n\n".to_string(),
            "Authority and Requirements" => "- R1: y\n\n".to_string(),
            "Non-goals" => "- NG1: z\n\n".to_string(),
            "Approach and Risks" => "- approach: a\n\n".to_string(),
            "Phases and Verification" => {
                "- phases:\n  - phase_slug: `p1`\n    - story_id: `p1-1`\n    - status: checked\n    - goal: R1\n\n"
                    .to_string()
            }
            "Progress" => "- 2026-09-19T08:00Z — p1 — task_status=DONE — ok\n\n".to_string(),
            "Decisions" => "- 2026-09-19 — p1 — chose a\n\n".to_string(),
            "Validation" => format!("{LEGACY_VALIDATION}\n"),
            "Current State and Next Action" => {
                "- active_phase: p1\n- lifecycle_status: checked\n".to_string()
            }
            other => panic!("no body for {other}"),
        }
    }

    /// A 9-section plan with the given heading order.
    fn legacy_plan(order: &[&str]) -> String {
        let mut s = String::from(
            "---\nid: p-1\nintake_id: i-1\nlane: normal\nstatus: active\n---\n\n# Plan: p\n\n",
        );
        for h in order {
            s.push_str(&format!("## {h}\n{}", legacy_body(h)));
        }
        s
    }

    fn headings_of(s: &str) -> Vec<String> {
        s.split('\n')
            .filter(|l| l.starts_with("## "))
            .map(|l| l.to_string())
            .collect()
    }

    /// The legacy 9-section plan migrates to the 5-section one in either
    /// heading order, drops `intake_id`/`story_id` and each phase's
    /// `- status:` bullet, copies `## Validation` verbatim, and is idempotent.
    #[test]
    fn migrate_plan_legacy_to_five_sections() {
        let reordered = [
            "Outcome",
            "Authority and Requirements",
            "Non-goals",
            "Approach and Risks",
            "Phases and Verification",
            "Current State and Next Action",
            "Validation",
            "Decisions",
            "Progress",
        ];
        for (name, order) in [
            ("canonical order", LEGACY_PLAN_SECTIONS.to_vec()),
            ("current state before validation", reordered.to_vec()),
        ] {
            let input = legacy_plan(&order);
            let (out, changed) =
                migrate_plan(input.as_bytes()).unwrap_or_else(|e| panic!("{name}: {e}"));
            assert!(changed, "{name}: migrate_plan reported no change");
            let got = String::from_utf8(out.clone()).expect("migrated plan is UTF-8");

            let want = [
                "## Goal",
                "## Phases and Verification",
                "## Log",
                "## Validation",
                "## Current State and Next Action",
            ];
            assert_eq!(
                headings_of(&got).join("|"),
                want.join("|"),
                "{name}: headings"
            );
            for s in [
                "### Authority and Requirements\n- R1: y",
                "### Non-goals\n",
                "### Decisions\n- 2026-09-19 — p1 — chose a",
                "## Phases and Verification\n- approach: a\n\n- phases:",
                "  - phase_slug: `p1`\n    status: checked\n",
            ] {
                assert!(got.contains(s), "{name}: missing {s:?} in:\n{got}");
            }
            assert!(
                got.contains(&format!("## Validation\n{LEGACY_VALIDATION}")),
                "{name}: Validation body not copied verbatim"
            );
            for s in ["intake_id", "story_id", "- status:"] {
                assert!(!got.contains(s), "{name}: still contains {s:?}");
            }
            assert!(
                got.starts_with(
                    "---\nid: p-1\nlane: normal\nstatus: active\n---\n\n# Plan: p\n\n## Goal\n"
                ),
                "{name}: frontmatter or preamble changed:\n{got}"
            );

            let (_, _, before) = split_plan(&input).expect("input splits");
            let (_, _, after) = split_plan(&got).expect("output splits");
            assert_eq!(
                sha(before["Validation"].as_bytes()),
                sha(after["Validation"].as_bytes()),
                "{name}: Validation bytes changed"
            );

            let (again, changed) = migrate_plan(&out).unwrap_or_else(|e| panic!("{name}: {e}"));
            assert!(
                !changed && again == out,
                "{name}: second run changed the plan"
            );
        }
    }

    /// Any heading set that is not exactly the legacy nine is returned
    /// unchanged, so an unrecognized plan is never rewritten.
    #[test]
    fn migrate_plan_unknown_sets_untouched() {
        let mut dup = LEGACY_PLAN_SECTIONS.to_vec();
        dup.push("Progress");
        let canonical = legacy_plan(&LEGACY_PLAN_SECTIONS);
        let cases = [
            ("missing section", legacy_plan(&LEGACY_PLAN_SECTIONS[..8])),
            ("extra heading", format!("{canonical}## Notes\nx\n")),
            ("duplicate heading", legacy_plan(&dup)),
            (
                "renamed heading",
                canonical.replace("## Current State and Next Action", "## Current State"),
            ),
            ("no headings", "# plan a".to_string()),
        ];
        for (name, input) in cases {
            let (out, changed) =
                migrate_plan(input.as_bytes()).unwrap_or_else(|e| panic!("{name}: {e}"));
            assert!(
                !changed && out == input.as_bytes(),
                "{name}: migrate_plan changed an unknown plan"
            );
        }
    }

    /// `update` migrates the one active plan and reports it; a second update
    /// leaves it alone.
    #[test]
    fn update_migrates_active_plan() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let plan = "docs/plans/active/p.md";
        write_file(root, plan, &legacy_plan(&LEGACY_PLAN_SECTIONS));

        let out = update_ok(root);
        assert!(
            out.contains(&format!("migrated       {plan}")),
            "no migrated line:\n{out}"
        );
        assert!(
            read_file(root, plan).contains("## Log\n"),
            "plan not migrated"
        );

        let out = update_ok(root);
        assert!(
            !out.contains("migrated") && !out.contains("notice"),
            "second update touched the plan:\n{out}"
        );
    }

    /// An active plan whose section set is unrecognized is left byte-identical
    /// and reported with a notice.
    #[test]
    fn update_unknown_plan_left_with_notice() {
        let _env = IsolatedEnv::new();
        let repo = temp_repo();
        let root = repo.path();
        install_ok(root);
        let plan = "docs/plans/active/p.md";
        let input = format!("{}## Notes\nx\n", legacy_plan(&LEGACY_PLAN_SECTIONS));
        write_file(root, plan, &input);

        let out = update_ok(root);
        assert!(
            out.contains(&format!(
                "notice     {plan} has an unrecognized section set"
            )),
            "no notice:\n{out}"
        );
        assert_eq!(read_file(root, plan), input, "unknown plan was modified");
    }
}
