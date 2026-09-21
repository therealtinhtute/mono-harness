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
