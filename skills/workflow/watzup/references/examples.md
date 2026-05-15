# Watzup Recap Examples

All examples conform to `output-contract.md`. None contain skill-internal mechanics.

---

## Example 1: Resume with WIP and committed work

**Scenario:** Start of session on `feature/add-auth`. Branch has committed work and uncommitted WIP. HANDOFF.md exists.

**Input:**
```
/watzup
```

**Output:**
```
Recap — feature/add-auth (2026-05-15)

Trạng thái
- Nhánh: feature/add-auth, 5 commits ahead of main
- Uncommitted: 3 files, +45/-12 lines
- Readiness: needs-work

Context
- Handoff: JWT refresh flow partially implemented, stopped at token revocation
- Artifact chain: skipped

Thay đổi
- Hệ thống xác thực login/logout/session
- JWT token handling và middleware guards
- [WIP] Token revocation endpoint — chưa hoàn thành

Next: Hoàn thành token revocation trong src/auth/revoke.ts rồi bổ sung tests.
```

---

## Example 2: Clean branch ready for PR

**Scenario:** All work committed, clean tree, no blockers.

**Input:**
```
/watzup
```

**Output:**
```
Recap — feature/add-auth (2026-05-15)

Trạng thái
- Nhánh: feature/add-auth, 8 commits ahead of main
- Working tree: sạch
- Readiness: ready-for-pr

Context
- Không có handoff

Thay đổi
- Hệ thống xác thực end-to-end (login, logout, JWT, session)
- Middleware và route guards cho API endpoints
- Integration tests cho luồng auth

Next: Chạy /check review rồi tạo PR vào main.
```

---

## Example 3: WIP with risks

**Scenario:** Database refactor with uncommitted migrations and quality issues.

**Input:**
```
/watzup
```

**Output:**
```
Recap — feature/refactor-db (2026-05-15)

Trạng thái
- Nhánh: feature/refactor-db, 3 commits ahead of main
- Uncommitted: 2 files, +180/-0 lines
- Readiness: needs-work

Context
- Không có handoff

Thay đổi
- Schema migration (users, sessions, tokens)
- Cập nhật ORM models
- [WIP] Migration scripts — chưa có rollback

Risks
| Risk | Mức độ | Action |
|------|--------|--------|
| Schema migration không có rollback | cao | Thêm rollback migration trước khi commit |
| Thiếu integration test cho migration | vừa | Thêm test trên dataset mẫu |

Next: Bổ sung rollback migration cho 0042_*.sql.
```

---

## Example 4: Empty state

**Scenario:** Clean tree, no commits ahead of main, no handoff.

**Input:**
```
/watzup
```

**Output:**
```
Nhánh sạch — không có thay đổi nào so với main.
Next: Bắt đầu task mới hoặc kéo thay đổi mới nhất.
```

---

## Example 5: Resume with harness artifacts

**Scenario:** Midway through a cook pipeline. HANDOFF.md and artifacts exist.

**Input:**
```
/watzup
```

**Output:**
```
Recap — feature/inbox-ui (2026-05-15)

Trạng thái
- Nhánh: feature/inbox-ui, 4 commits ahead of main
- Working tree: sạch
- Readiness: needs-work

Context
- Handoff: Phase inbox-ui hoàn tất cook, gate còn thiếu proof cho 1 task
- Phase: inbox-ui | Cook run: passed | Check: needs-proof

Thay đổi
- Inbox UI component với list/detail views
- API integration cho inbox endpoints
- Unit tests cho inbox components

Risks
| Risk | Mức độ | Action |
|------|--------|--------|
| Proof trail thiếu cho inbox-ui task | vừa | Bổ sung verification output rồi chạy lại /check |

Next: Cập nhật run artifact rồi chạy /check gate.
```

---

## Example 6: Blocked state from handoff

**Scenario:** Previous session ended with explicit blocker.

**Input:**
```
/watzup
```

**Output:**
```
Recap — feature/triage-rules (2026-05-15)

Trạng thái
- Nhánh: feature/triage-rules, 2 commits ahead of main
- Uncommitted: 1 file
- Readiness: blocked

Context
- Handoff: Dừng ở phase triage-rules — spec và plan lệch nhau (contract drift)
- Phase: triage-rules | Cook run: blocked

Thay đổi
- Triage rule engine cơ bản
- [WIP] Rule configuration — dừng do scope conflict

Risks
| Risk | Mức độ | Action |
|------|--------|--------|
| Spec và phase plan lệch nhau | cao | Chạy /plan phase triage-rules để khóa scope đúng |

Next: Chạy /plan phase triage-rules để refresh scope trước khi code tiếp.
```
