# Harness Retrospective Examples

## Example 1 — Proof gap after check

**Output:**
```text
Session — feature/inbox-ui (2026-05-11)

Trạng thái:
- Nhánh: feature/inbox-ui
- Vị trí: 4 commits ahead of main
- Working tree: sạch
- Readiness: needs-proof

Thay đổi chính:
- Hoàn tất phase inbox-ui và có run log cho cook
- Gate đã chạy nhưng còn thiếu proof cho 1 task verification

Risks:
| Risk | Mức độ | Action |
|------|--------|--------|
| Proof trail thiếu cho inbox-ui | vừa | Bổ sung verification output rồi chạy lại check |

Next: Cập nhật run artifact của inbox-ui rồi chạy lại phase gate.
```

## Example 2 — Plan refresh needed

**Output:**
```text
Session — feature/triage-rules (2026-05-11)

Trạng thái:
- Nhánh: feature/triage-rules
- Vị trí: 2 commits ahead of main
- Working tree: 1 file uncommitted
- Readiness: needs-plan-refresh

Thay đổi chính:
- Triển khai dừng ở phase triage-rules do lệch giữa spec và plan
- Handoff đã ghi rõ blocker theo contract drift

Risks:
| Risk | Mức độ | Action |
|------|--------|--------|
| Spec và phase plan lệch nhau | cao | Refresh phase plan trước khi code tiếp |

Next: Chạy lại plan cho triage-rules để khóa scope đúng.
```
