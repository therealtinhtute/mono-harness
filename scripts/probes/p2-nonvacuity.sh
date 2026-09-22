#!/bin/bash
# Re-verify that the three refuses_symlink tests fail against the PRE-FIX code.
# Reverts only the fix hunks (keeping the new tests), runs them, then restores
# both files and proves the restore was byte-identical.
set -u
cd "$(git rev-parse --show-toplevel)" || exit 1
MOD=cli/src/installer/mod.rs
UNI=cli/src/installer/uninstall.rs
cp "$MOD" /tmp/nv-mod.bak
cp "$UNI" /tmp/nv-uni.bak
before_mod=$(sha256sum "$MOD" | cut -d' ' -f1)
before_uni=$(sha256sum "$UNI" | cut -d' ' -f1)
restore() { cp /tmp/nv-mod.bak "$MOD"; cp /tmp/nv-uni.bak "$UNI"; }
trap restore EXIT

python3 - <<'PYEOF'
mod, uni = 'cli/src/installer/mod.rs', 'cli/src/installer/uninstall.rs'
new_fn = '''    let tmp = PathBuf::from(format!("{}.tmp-zharness", p.display()));
    match fs::remove_file(&tmp) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(e) => return Err(format!("write {}: {e}", tmp.display())),
    }
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|e| format!("write {}: {e}", tmp.display()))?;
    f.write_all(data)
        .map_err(|e| format!("write {}: {e}", tmp.display()))?;
    drop(f);
'''
old_fn = '''    let tmp = PathBuf::from(format!("{}.tmp-zharness", p.display()));
    fs::write(&tmp, data).map_err(|e| format!("write {}: {e}", tmp.display()))?;
'''
s = open(mod).read()
assert s.count(new_fn) == 1, 'write_file_atomic hunk not found'
s = s.replace(new_fn, old_fn)
assert s.count('write_file_atomic(&ap, body.as_bytes())?;') == 1
s = s.replace('write_file_atomic(&ap, body.as_bytes())?;',
              'fs::write(&ap, body).map_err(|e| format!("write {}: {e}", ap.display()))?;')
open(mod, 'w').write(s)

t = open(uni).read()
new_r = '''                // Through the atomic writer: a symlink planted at the managed
                // path must be replaced, never followed.
                let _ = write_file_atomic(&dst_p, o);'''
assert t.count(new_r) == 1, 'uninstall restore hunk not found'
t = t.replace(new_r, '                let _ = fs::write(&dst_p, o);')
open(uni, 'w').write(t)
print('fix hunks reverted (tests kept)')
PYEOF

echo "--- pre-fix test run (expect 3 FAILED) ---"
(cd cli && cargo test --lib refuses_symlink 2>&1 | grep -E '^test |^test result')
restore
trap - EXIT

echo "--- restore check ---"
after_mod=$(sha256sum "$MOD" | cut -d' ' -f1)
after_uni=$(sha256sum "$UNI" | cut -d' ' -f1)
[ "$before_mod" = "$after_mod" ] && echo "mod.rs restored byte-identically" || echo "mod.rs NOT RESTORED"
[ "$before_uni" = "$after_uni" ] && echo "uninstall.rs restored byte-identically" || echo "uninstall.rs NOT RESTORED"
echo "--- post-restore test run (expect 3 ok) ---"
(cd cli && cargo test --lib refuses_symlink 2>&1 | grep -E '^test result')
