# Watzup Skill Examples

All examples conform to `output-contract.md`. None contain skill-internal mechanics.

---

## Example 1: Standard fast review

**Scenario:** Wrap up a session on `feature/add-auth` after committed feature work.

**Input:**
```
/watzup
```

**Output:**
```
Session — feature/add-auth (2026-05-09)

Trạng thái:
- Nhánh: feature/add-auth
- Vị trí: 5 commits ahead of main
- Working tree: sạch

Thay đổi chính:
- Thêm hệ thống xác thực (login, logout, session)
- Triển khai xử lý JWT token
- Bổ sung middleware và guards cho route bảo vệ

Next: Bổ sung integration test cho luồng auth, sau đó tạo PR.
```

---

## Example 2: Pre-PR fast review

**Scenario:** Branch is committed and tested; verify PR readiness.

**Input:**
```
/watzup feature/add-auth
```

**Output:**
```
Session — feature/add-auth (2026-05-09)

Trạng thái:
- Nhánh: feature/add-auth
- Vị trí: 8 commits ahead of main
- Working tree: sạch

Thay đổi chính:
- Hệ thống xác thực hoàn chỉnh (login, logout, session, JWT)
- Middleware và route guards
- Integration test cho luồng auth
- Cập nhật tài liệu API

Next: Tạo PR vào main.
```

---

## Example 3: Fast review with risks

**Scenario:** Database refactor branch with multiple safety gaps.

**Input:**
```
/watzup
```

**Output:**
```
Session — feature/refactor-db (2026-05-09)

Trạng thái:
- Nhánh: feature/refactor-db
- Vị trí: 3 commits ahead of main
- Working tree: 2 file uncommitted (migrations/0042_*.sql)

Thay đổi chính:
- Schema migration (users, sessions, tokens)
- Cập nhật ORM model
- Migration scripts

Risks:
| Risk | Mức độ | Action |
|------|--------|--------|
| Schema migration không có rollback | cao | Thêm rollback migration trước khi merge |
| Thay đổi schema lớn, có thể breaking | cao | Test trên staging trước khi tạo PR |
| Thiếu integration test cho migration | vừa | Thêm test trên dataset mẫu |

Next: Bổ sung rollback migration và staging test trước khi tạo PR.
```

---

## Example 4: Fast clean state

**Scenario:** Working tree clean, no new activity since last review.

**Input:**
```
/watzup
```

**Output:**
```
Đã sạch — không có thay đổi.
Next: Bắt đầu nhánh mới hoặc kéo thay đổi mới nhất từ remote.
```

No file is written even if invoked as `/watzup deep` in this state.

---

## Example 5: Deep mode

**Scenario:** Pre-PR comprehensive review on `feature/add-auth`. Markdown by default; HTML excerpt below shows the self-contained format.

**Input:**
```
/watzup feature/add-auth deep
```

**File written:** `.kit/reports/watzup/20260509-feature-add-auth.md`

**File content:**
```markdown
---
title: Session — feature/add-auth (2026-05-09)
branch: feature/add-auth
commits: 8
files: 18
created: 2026-05-09
tags: [watzup, review, session]
---

# Session — feature/add-auth (2026-05-09)

## Trạng thái
- Nhánh: feature/add-auth
- Vị trí: 8 commits ahead of main
- Working tree: sạch

## Changes Overview
- Commits: 8 (feat: 5, test: 2, docs: 1)
- Files: 14 modified, 4 added, 0 removed
- Lines: +680 / -95

## Key Changes
1. Hệ thống xác thực end-to-end — bao phủ login, logout, session, JWT
2. Middleware và route guards — bảo vệ API endpoints
3. Integration test cho luồng auth — coverage tăng đáng kể
4. Cập nhật tài liệu API — endpoints mới được mô tả đầy đủ

## Quality Assessment
- Test Coverage: increased
- Documentation: updated
- Breaking Changes: no

## Risks & Blockers
| Risk | Mức độ | Action |
|------|--------|--------|
| Token refresh chưa có rate limit | vừa | Thêm middleware giới hạn tần suất refresh trước khi merge |
| Logout không revoke refresh token phía server | thấp | Thêm endpoint revoke và gọi từ client khi logout |

## Next Steps
1. Áp dụng rate limit cho token refresh
2. Bổ sung revoke flow phía server
3. Tạo PR vào main và mời reviewer cho phần JWT handling
```

**Console summary** (printed alongside the file): same shape as Example 2.

---

### HTML output excerpt (`/watzup feature/add-auth deep --format=html`)

```html
<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8">
  <title>Session — feature/add-auth (2026-05-09)</title>
  <style>
    body { font-family: -apple-system, sans-serif; max-width: 720px; margin: 2rem auto; line-height: 1.5; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border: 1px solid #ddd; padding: 6px 10px; text-align: left; }
    th { background: #f6f8fa; }
  </style>
</head>
<body>
  <h1>Session — feature/add-auth (2026-05-09)</h1>
  <!-- Trạng thái, Changes Overview, Key Changes, Quality Assessment, Risks & Blockers, Next Steps follow -->
</body>
</html>
```

The HTML file is fully self-contained: no `<link rel="stylesheet">`, no `<script src="...">`, no external fonts or CDN URLs.
