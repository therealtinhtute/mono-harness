//! The canonical doc set bundled into the binary at compile time.
//!
//! The embedded set mirrors `cli/docs/embedded/embed.go` exactly:
//! `AGENTS.md`, `WORKFLOW.md`, `playbooks/`, and — separately, as Go's second
//! `embed.FS` — `templates/`. The Go sources that live in that directory
//! (`embed.go`, `parity_test.go`) are deliberately not embedded: embedding the
//! directory whole would ship them in the binary and put them in the manifest.

use include_dir::{include_dir, Dir};

static PLAYBOOKS: Dir<'static> = include_dir!("$CARGO_MANIFEST_DIR/docs/embedded/playbooks");
static TEMPLATES: Dir<'static> = include_dir!("$CARGO_MANIFEST_DIR/docs/embedded/templates");

const AGENTS_MD: &[u8] = include_bytes!(concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/docs/embedded/AGENTS.md"
));
const WORKFLOW_MD: &[u8] = include_bytes!(concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/docs/embedded/WORKFLOW.md"
));

/// The write-once project identity template, written to `docs/PROJECT.md` only
/// when that file is absent.
pub const PROJECT_TEMPLATE: &str = "templates/project.identity.md";

/// Every embedded doc path plus the docs version stamp identifying the build.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Manifest {
    pub paths: Vec<String>,
    pub docs_version: String,
}

/// Read one embedded file by its path inside the doc set. The `templates/`
/// prefix is kept, mirroring Go's `Templates.ReadFile(t.Src)`.
pub fn read_file(path: &str) -> Option<&'static [u8]> {
    match path {
        "AGENTS.md" => Some(AGENTS_MD),
        "WORKFLOW.md" => Some(WORKFLOW_MD),
        _ => {
            if let Some(rest) = path.strip_prefix("playbooks/") {
                PLAYBOOKS.get_file(rest).map(|f| f.contents())
            } else if let Some(rest) = path.strip_prefix("templates/") {
                TEMPLATES.get_file(rest).map(|f| f.contents())
            } else {
                None
            }
        }
    }
}

/// The canonical playbook file names (six stages plus the `work-full` and
/// `check-validation` companions), sorted.
pub fn playbook_names() -> Vec<String> {
    let mut names: Vec<String> = PLAYBOOKS.files().map(|f| file_name(f.path())).collect();
    names.sort();
    names
}

/// How many playbook files are embedded.
pub fn playbook_count() -> usize {
    PLAYBOOKS.files().count()
}

/// Every embedded path in the managed set, sorted — the same list Go's
/// `BuildManifest` walks out of `FS`. `templates/` is excluded: it is the
/// separate embed, never walked into the manifest.
pub fn managed_paths() -> Vec<String> {
    let mut paths = vec!["AGENTS.md".to_string(), "WORKFLOW.md".to_string()];
    for name in playbook_names() {
        paths.push(format!("playbooks/{name}"));
    }
    paths.sort();
    paths
}

/// Every embedded path across both embeds, sorted.
pub fn all_embedded_paths() -> Vec<String> {
    let mut paths = managed_paths();
    for f in TEMPLATES.files() {
        paths.push(format!("templates/{}", file_name(f.path())));
    }
    paths.sort();
    paths
}

/// Walk the doc set and return every managed path, sorted, plus the supplied
/// docs version (the CLI's own version string).
pub fn build_manifest(docs_version: &str) -> Manifest {
    Manifest {
        paths: managed_paths(),
        docs_version: docs_version.to_string(),
    }
}

fn file_name(p: &std::path::Path) -> String {
    p.file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_default()
}

#[cfg(test)]
mod tests {
    use std::fs;
    use std::path::{Path, PathBuf};

    use super::*;

    /// Folded at compile time so a tree scan cannot match its own guard list.
    const LEGACY_DB_NAME: &str = concat!("harness", ".db");

    #[test]
    fn build_manifest_paths_present_and_non_empty() {
        let m = build_manifest("dev");
        assert!(!m.paths.is_empty(), "manifest has no paths");
        for p in &m.paths {
            let data = read_file(p).unwrap_or_else(|| panic!("read_file({p}): not embedded"));
            assert!(!data.is_empty(), "{p} is empty");
        }
    }

    #[test]
    fn playbook_count_is_eight() {
        assert_eq!(playbook_count(), 8);
    }

    #[test]
    fn build_manifest_entrypoint_and_workflow_present() {
        let m = build_manifest("dev");
        for w in ["AGENTS.md", "WORKFLOW.md"] {
            assert!(m.paths.iter().any(|p| p == w), "manifest missing {w}");
        }
    }

    #[test]
    fn build_manifest_docs_version_exposed() {
        assert_eq!(build_manifest("0.2.0").docs_version, "0.2.0");
    }

    /// Every file on disk under `docs/embedded` must be embedded through
    /// either embed — the managed set walked into the manifest or the
    /// separately embedded `templates/` — and vice versa. A doc added on disk
    /// but not to an embed (or the reverse) is invisible to the binary.
    #[test]
    fn build_manifest_matches_disk_tree() {
        let manifest_paths = all_embedded_paths();

        let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("docs/embedded");
        let mut disk_paths = Vec::new();
        collect(&root, &root, &mut disk_paths);
        // embed.go carries the directive; it is not itself embedded.
        disk_paths.retain(|p| !p.ends_with(".go"));
        disk_paths.sort();

        assert_eq!(
            manifest_paths, disk_paths,
            "manifest and disk tree disagree\nmanifest: {manifest_paths:?}\ndisk: {disk_paths:?}"
        );
    }

    fn collect(root: &Path, dir: &Path, out: &mut Vec<String>) {
        let entries = fs::read_dir(dir).unwrap_or_else(|e| panic!("walk {}: {e}", dir.display()));
        for e in entries.flatten() {
            let p = e.path();
            if p.is_dir() {
                collect(root, &p, out);
            } else {
                out.push(
                    p.strip_prefix(root)
                        .unwrap_or(&p)
                        .to_string_lossy()
                        .replace('\\', "/"),
                );
            }
        }
    }

    struct PlaybookContract {
        name: &'static str,
        path: &'static str,
        required: &'static [&'static str],
        forbidden: &'static [&'static str],
    }

    /// Strings that must never reappear in any embedded playbook: the retired
    /// lifecycle verbs and the pre-v0.15 state-mirroring vocabulary.
    const RETIRED: &[&str] = &[
        ".kit/planning",
        "SPEC.md",
        "ROADMAP.md",
        "-CONTEXT.md",
        "-PLAN.md",
        ".kit/runs",
        ".kit/reports",
        ".kit/HANDOFF.md",
        "implementation-notes",
        "--artifact-path",
        "zharness scaffold run",
        "zharness scaffold check",
        "zharness scaffold handoff",
        "RUN artifact",
        "CHECK report",
        // v0.15 slim (R1/R6): every lifecycle verb is deleted from source, and
        // the playbooks are markdown-first with no index-sync block. These
        // strings must never reappear in any embedded playbook.
        "zharness preflight",
        "zharness run create",
        "zharness trace add",
        "zharness decision add",
        "zharness story",
        "zharness intake",
        "zharness id",
        "zharness scaffold plan",
        "zharness query",
        "zharness audit",
        "zharness validate",
        "zharness resume",
        "zharness init",
        "zharness migrate",
        "zharness import",
        "zharness db ",
        "zharness memory",
        "zharness plan complete",
        "zharness handoff record",
        "check record",
        "--close-phase",
        "Optional index-sync",
        LEGACY_DB_NAME,
        "lifecycle ledger",
        "DB-mirroring",
        "mirrored check row",
        "check_id: ULID",
        "latest_run_id",
        "lifecycle rows",
    ];

    #[test]
    fn one_plan_playbook_contract() {
        let contracts = [
            PlaybookContract {
                name: "brainstorm locks honest bootstrap state",
                path: "playbooks/brainstorm.md",
                required: &[
                    "## Goal",
                    "## Phases and Verification",
                    "## Log",
                    "## Validation",
                    "## Current State and Next Action",
                    "The canonical active path is `docs/plans/active/{slug}.md`",
                    "confirm no non-empty plan exists under `docs/plans/active/`",
                    "explore creates no plans, reports, changesets, or markdown artifacts",
                    "approach: not-planned",
                    "exact_next_action: to-plan",
                    "`tiny` → stop and route to `work bounded`",
                    "- success_signal:",
                    "- actors:",
                    "| acceptance: <how it is verified> |",
                    "those definitions are immutable",
                    "`## Log` is the sole task execution-status source",
                ],
                forbidden: &[
                    "story_id",
                    "intake_id",
                    "## Progress",
                    "## Decisions",
                    "new-spec",
                ],
            },
            PlaybookContract {
                name: "to-plan defines phases as markdown truth",
                path: "playbooks/to-plan.md",
                required: &[
                    "## Phases and Verification",
                    "- approach:",
                    "indented two spaces with no bullet",
                    "— output: <expected result> — check: `<command>` — stop_if: <condition>",
                    "- escalate_when:",
                    "every requirement appears in some phase `goal:`",
                    "validation expectations",
                    "docs/plans/active/{slug}.md",
                    "After a phase/task definition is written, it is immutable",
                    "work/check/handoff may change only that phase's lifecycle status in this file",
                    "Do not add task status fields",
                    "never write parallel task/phase state anywhere else",
                    "`## Log` is the sole task execution-status source",
                ],
                forbidden: &["story_id", "## Approach and Risks", "## Progress"],
            },
            PlaybookContract {
                name: "work routes full mode to its companion",
                path: "playbooks/work.md",
                required: &[
                    "read `docs/playbooks/work-full.md` now",
                    "print the resolved mode as your first output line",
                    "bounded/simple mode creates no plans, reports, changesets, or markdown artifacts",
                ],
                forbidden: &[],
            },
            PlaybookContract {
                name: "work appends durable markdown log entries",
                path: "playbooks/work-full.md",
                required: &[
                    "## Log",
                    "## Current State and Next Action",
                    "slice `docs/plans/active/{slug}.md` by section",
                    "set that phase's plan status to `in-progress`",
                    "`start|done|blocked|decision`",
                    "A `blocked` entry is written immediately",
                    "Do not add or update task-definition `status` fields",
                    "`## Log` is the sole task execution-status source",
                ],
                forbidden: &["## Progress", "## Decisions", "task_status="],
            },
            PlaybookContract {
                name: "check preserves review intent and records evidence",
                path: "playbooks/check.md",
                required: &[
                    "## Validation",
                    "Invocation intent wins",
                    "`auto` never resolves to `full`",
                    // These pin the authored routing contract, not agent behavior.
                    "resolve to `bounded` without reading any plan",
                    "invalid initiative state never falls back to `bounded`",
                    "count every non-empty markdown plan under `docs/plans/active/` before matching the requested initiative",
                    "Require exactly one active plan in total",
                    "stop even when only one matches the request",
                    "Read only its selected phase and `## Current State and Next Action`",
                    "re-reading after any compaction",
                    "including unstarted, `checked`, or `done`",
                    "stop before checks or writes",
                    "`bounded` (aliases: `simple`, `review`)",
                    "After the preflight succeeds, print the resolved mode before running checks or writing state",
                    "A failed preflight reports its blocker instead",
                    "`bounded` is always response-only",
                    "never appends to Validation",
                    "set the phase status and Current State lifecycle status to `checked`",
                    "keep the phase and Current State lifecycle status `in-progress`",
                    "Durable `gate` runs automated checks",
                    "`full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review",
                    "`gate` does not perform that complete manual review",
                    "read `docs/playbooks/check-validation.md`",
                    "The repository's pre-commit hook is the sole proof guarantee",
                    "REQUEST_CHANGES entries may cite deliberately failing commands",
                    "`## Log` is the sole task execution-status source",
                    "record it in the entry's `requirements:` line",
                    "`not met` is a material plan contradiction",
                    "also judges the Goal's `success_signal:`",
                    "a `requirements:` line is testimony",
                    "Record `rollback_point:`",
                    "depth: quick | standard | deep",
                ],
                forbidden: &[
                    "## Progress",
                    "receipt:",
                    "mode: gate | full | review",
                    "Before reading any plan, print the resolved mode",
                    "whose selected phase reads `in-progress`; otherwise to `bounded`",
                    "Durable `gate`/`full` mode runs real checks and review",
                    "Gate/full: applicable commands have captured output, alignment and code review ran",
                ],
            },
            PlaybookContract {
                name: "check-validation defines guard-visible evidence",
                path: "playbooks/check-validation.md",
                required: &[
                    "Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output",
                    "so the commit-time guard can find them",
                    "The first line carries the anchored token `verdict: <VERDICT>`",
                    "at most 3 lines",
                    "- requirements: R1 met (<proof>) | R2 not met (<gap>)",
                    "- rollback_point:",
                    "Validation is append-only from an entry's first commit",
                    "**Proof re-execution contract**",
                ],
                forbidden: &[],
            },
            PlaybookContract {
                name: "handoff closes phases before initiatives",
                path: "playbooks/handoff.md",
                required: &[
                    "## Current State and Next Action",
                    "Close every cleanly checked phase",
                    "keep frontmatter `status: active` and the same active path",
                    "Before closing the final phase, require every prior phase to be `done`",
                    "Only after the plan shows every phase `done`",
                    "`lifecycle_status: completed`",
                    "git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md",
                    "exactly one file may represent the initiative afterwards",
                    "docs/plans/active/{slug}.md",
                    "docs/plans/completed/{slug}.md",
                    "Preserve every phase/task definition",
                    "`## Log` is the sole task execution-status source",
                    "drop `start` and `done` entries",
                    "keep the last entry per phase byte-identical",
                ],
                forbidden: &[
                    "## Progress",
                    "## Decisions",
                    "zharness preflight handoff --mode full --json",
                    "Require every phase to be `done` or the final phase",
                ],
            },
            PlaybookContract {
                name: "watzup recaps without writing",
                path: "playbooks/watzup.md",
                required: &[
                    "Select the active plan by name",
                    "exactly one non-empty file may exist under `docs/plans/active/*.md`",
                    "Remain read-only",
                    "docs/plans/active/{slug}.md",
                    "never the whole file",
                    "Read task execution status only from append-only `## Log`",
                    "the sole task execution-status source",
                ],
                forbidden: &["zharness resume --json"],
            },
        ];

        for c in &contracts {
            let data =
                read_file(c.path).unwrap_or_else(|| panic!("read_file({}): not embedded", c.path));
            let content = String::from_utf8_lossy(data);
            for phrase in c.required {
                assert!(
                    content.contains(phrase),
                    "{} ({}): missing one-plan contract phrase {phrase:?}",
                    c.path,
                    c.name
                );
            }
            for phrase in RETIRED.iter().chain(c.forbidden.iter()) {
                assert!(
                    !content.contains(phrase),
                    "{} ({}): contains forbidden lifecycle contract {phrase:?}",
                    c.path,
                    c.name
                );
            }
        }
    }

    /// Live instructions must not route agents to the retired `.kit/`
    /// directory; per-machine scratch lives under `.zharness/cache/`. `docs/`
    /// is excluded because it also holds historical records that describe
    /// `.kit/` accurately.
    #[test]
    fn no_legacy_kit_paths() {
        let repo_root = Path::new(env!("CARGO_MANIFEST_DIR"))
            .parent()
            .expect("crate lives one level below the repository root");
        for rel in [
            "skills",
            "rules",
            "scripts",
            "setup",
            "site",
            "README.md",
            "AGENTS.md",
        ] {
            let p = repo_root.join(rel);
            assert!(p.exists(), "walk {rel}: no such path");
            let mut hits: Vec<PathBuf> = Vec::new();
            scan_for_kit(&p, &mut hits);
            assert!(
                hits.is_empty(),
                "{hits:?} reference retired .kit/; use .zharness/cache/ or docs/"
            );
        }
    }

    fn scan_for_kit(p: &Path, hits: &mut Vec<PathBuf>) {
        if p.is_dir() {
            let entries = fs::read_dir(p).unwrap_or_else(|e| panic!("walk {}: {e}", p.display()));
            for e in entries.flatten() {
                scan_for_kit(&e.path(), hits);
            }
            return;
        }
        let Ok(data) = fs::read(p) else {
            return;
        };
        if String::from_utf8_lossy(&data).contains(".kit/") {
            hits.push(p.to_path_buf());
        }
    }
}
