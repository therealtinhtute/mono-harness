# Interview record: closing the v0.15 blockers

> **Archive — the `/interview` pass run 2026-08-26** between the review
> (`docs/references/zharness-v015/review-findings.md`) and `to-plan`. Seven questions, each
> answered by the owner. Every answer was the recommended option.
>
> This is the decision record that turned the pre-review draft
> (`docs/references/zharness-v015/v015-original-plan.md`) plus its superseded predecessor
> (`docs/references/zharness-v015/v013-plan.md`) into the single merged plan at
> `docs/plans/completed/zharness-v015-slim.md`.

---

## The seven questions


### Q1 · B3 SKILL.md

B3 — 6 file `skills/workflow/*/SKILL.md` trong repo này (mỗi file có 1 dòng STOP khi thiếu binary) sẽ xử lý thế nào? NG1 hiện nói "để nguyên, không nằm trong product path", nhưng đây chính là sản phẩm được cài qua `npx skills add`. Để nguyên = S1 (binary vắng mặt, zero STOP) không bao giờ pass.


| Option | What it meant |
|---|---|
| Sửa 6 SKILL.md trong repo, NG1 thu hẹp **← chosen** | NG1 chỉ còn cấm đụng bản đã cài ở ~/.claude/skills và ~/.agents/skills. Thêm 1 requirement mới: viết lại 6 SKILL.md của repo — bỏ STOP thiếu-binary + STOP version-gate, thay bằng 1 dòng degradation; sửa luôn skills/workflow/README.md:39,55. Đây là cách duy nhất S1 và S6 đạt được, và nó mở khoá luôn B1 (vì cùng file chứa `model:` pin). |
| Giữ NG1 nguyên, bỏ S1 và S6 | Chấp nhận binary vẫn là điều kiện cần để chạy skill. Plan không còn là "fail-open" đúng nghĩa — nó thành "markdown-only state" thuần tuý. Trung thực hơn nhưng mất phần lớn giá trị của hướng slim. |
| Sửa SKILL.md nhưng giữ version gate | Bỏ STOP khi thiếu binary (S1 pass), nhưng giữ cảnh báo version khi binary CÓ mà quá cũ. Nửa vời: vẫn còn 1 nhánh fail-closed, S6 phải đổi thành 3 guard. |


### Q2 · B1+B2 S7

B1+B2 — S7 (−30% cost, cache >80%) xử lý sao? Đòn bẩy lớn nhất (model boundary) đã dùng hết: audit dự báo P1–P3 → −31%, và cả 3 đều đã ship. Audit cũng nói thẳng phần v015 định làm (cắt playbook) chỉ đáng ~$0.002/phase.


| Option | What it meant |
|---|---|
| Xoá S7, thay bằng signal đo được của v015 **← chosen** | Đo thứ v015 thực sự thay đổi: token của một lần vào stage. Bỏ CLI = mất preflight packet (2,595 tok/stage) nhưng agent phải đọc markdown thô — có thể tăng. Đây mới là rủi ro thật v015 cần chứng minh, và đo được bằng byte không cần CLI. |
| Giữ S7 nhưng đổi baseline sang 0.14.0 | Thừa nhận −31% đã thu, đòi thêm −30% nữa từ $0.68 → $0.48. Cần thêm requirement bỏ `model:` pin (v013 P4). Nhưng audit:200 đã bác: brainstorm/to-plan cần opus, tiết kiệm không bù được rủi ro quyết định sai. |
| Giữ S7 nguyên, coi như đã pass | Ghi vào Validation rằng S7 đã đạt nhờ P1-P3 đã ship, trích bằng chứng work.md:41,45 + plan_query_test.go. Không thêm requirement nào. Nhanh nhưng S7 không còn gate được gì trong v015. |


### Q3 · B4+M1 guard

B4+M1 — cơ chế enforce của `record check` sau khi bỏ CLI. R2 hiện dựa vào "script pass marker" — nhưng agent vừa gọi script vừa viết entry, nên tự ghì marker được. Cả hai guard fail-closed đều treo ở đây.


| Option | What it meant |
|---|---|
| Hook tự re-execute, bỏ marker hẳn **← chosen** | Pre-commit hook đọc staged diff, tìm Validation entry MỚI, tự trích proof command từ sub-bullet và chạy lại — không tin bất kỳ marker nào agent viết. Cùng hook grep `lane:` frontmatter + `judge:` entry để chặn high-risk + same-session. Guarantee giữ nguyên như CONTRACT.md:189, không forgeable. |
| Giữ marker, thêm chữ ký nội dung | Marker = hash của (proof commands + diff sha). Hook verify hash khớp thay vì chạy lại — nhanh hơn, không phải chờ test. Nhưng agent vẫn tính được hash mà không chạy test thật. Chỉ chống nhầm, không chống cố ý. |
| Chỉ CI enforce, pre-commit chỉ cảnh báo | Pre-commit in warning, CI mới là cổng thật (chạy lại proof + check judge). Commit cục bộ không bị chặn. Đổi prevention lấy tốc độ — S6 phải sửa thành "2 detection", không còn fail-closed. |


### Q4 · M3+M4 scaffold

M3+M4 — ai scaffold, và scaffold những gì? R1 xoá `init`+`scaffold`, NG7 cấm `doctor` — nhưng `init` đang làm 5 việc: tạo DB, project WORKFLOW.md+playbooks, sửa AGENTS.md block, ghi .gitignore, track hash cho refresh. Bỏ DB thì 4 việc còn lại vẫn cần người làm. S5 còn đòi `docs/PROJECT.md` mà không R nào tạo.


| Option | What it meant |
|---|---|
| `install` = `init` trừ DB, thêm PROJECT.md **← chosen** | Installer nhận nguyên 4 việc còn lại của init, `.zharness/base/` thay `managed_docs` table làm nơi track hash. Thêm PROJECT.md template (K0, ≤50 dòng) vào managed set, brainstorm lock là bước ghi buộc trả lời. K1 (ARCHITECTURE) vẫn thuộc consumer — harness không nhân bản. |
| Installer chỉ lo file, PROJECT.md do skill lo | install/update chỉ quản playbooks + AGENTS.md + WORKFLOW.md. PROJECT.md do brainstorm tự tạo lần đầu khi lock, không nằm trong managed set → không bị updater đụng, hoàn toàn consumer-owned. Đơn giản hơn nhưng S5 phụ thuộc agent nhớ làm. |
| Bỏ S5, không làm PROJECT.md ở v0.15 | Tách knowledge architecture (K0-K5, greenfield/brownfield onboarding) ra initiative riêng sau v0.15. v0.15 chỉ tập trung: bỏ SQLite + installer-only + fail-open. Plan gọn hơn nhiều, ship nhanh hơn. |


### Q5 · M5+M6 playbook

M5+M6 — 6 playbook đang có 65 chỗ gọi `zharness` (work 19, handoff 16, check 11, brainstorm 8, to-plan 6, watzup 5) trên 391 dòng, hầu hết gọi lệnh R1 sắp xoá. Không R nào lo. Song song đó, R8 lấy `go test ./...` làm cổng mỗi phase — nhưng R1 xoá luôn tầng mà 68 file test / 12,770 dòng đang phủ.


| Option | What it meant |
|---|---|
| R15 viết lại playbook + đổi cổng sang kill-switch test **← chosen** | Playbook thuộc managed set (R12) nên phải viết lại cùng nhịp — thành 1 requirement rõ. Cổng mỗi phase đổi thành: kill-switch test chạy được + grep 0 tên lệnh đã chết trong docs/, thay vì dựa vào go test đang teo dần. |
| Viết lại playbook, giữ cổng như cũ | Thêm R15 cho playbook nhưng R8 giữ nguyên (go test + verify-doc-links). Chấp nhận cổng yếu dần theo từng phase, dựa vào kill-switch test cuối cùng ở S1. |
| Playbook viết lại thành 1 file duy nhất | Thay vì 6 playbook, gộp thành docs/WORKFLOW.md mở rộng (khuôn repository-harness). Ít file hơn, nhưng agent phải đọc cả file cho mọi stage — audit:202 nói playbook nằm trong cached prefix, đọc 1 lần/stage. |


### Q6 · M4 brownfield

M4 (phần còn lại) — brownfield: R10 đòi "deterministic detection → drafted proposal → human approval", nhưng v013 giao việc đó cho `doctor --adopt` và NG7 cấm `doctor`. Repo đã có sẵn README/CLAUDE.md/docs/** thì ai phát hiện và đề xuất?


| Option | What it meant |
|---|---|
| `install` tự nhận brownfield, in report read-only **← chosen** | Không thêm verb — giữ đúng install/update/uninstall của R1. `install` thấy repo đã có docs/plans hoặc CLAUDE.md → scaffold phần thiếu, KHÔNG đụng file sẵn có, in danh sách findings + đề xuất để người quyết. Khớp 0020 read-only-first. |
| `scripts/status.sh` làm adopt-detection | NG7 đã cho phép status.sh như tiện ích. Đẩy detection sang script repo-local — binary sạch tuyệt đối, chỉ install/update/uninstall. Nhưng script nằm TRONG repo đích, mà brownfield là lúc repo đó chưa có script. |
| Bỏ R10, brownfield sang initiative sau | v0.15 chỉ hỗ trợ greenfield + chính repo này (đã là brownfield thực tế, tự tay chuyển). Consumer có sẵn thì pin 0.14.x. Plan gọn hơn, nhưng R10 đang trích "issue #25 + 0020" — bỏ là mất bài học chống backfill giả. |


### Q7 · Phase order

Thứ tự phase và điểm dừng bắt buộc. Repo này đang tự chạy bằng harness sắp bị xoá — xoá CLI trước khi fail-open + hook guard sẵn sàng là tự đập chân mình, và R1 không quay đầu được.


| Option | What it meant |
|---|---|
| Fail-open → hook guard → PAUSE → xoá → installer → knowledge **← chosen** | Hai phase đầu không xoá gì, CLI vẫn nguyên, hoàn toàn revert được. Chứng minh kill-switch pass và hook thay được guarantee TRƯỚC khi đốt cầu. Pause bắt buộc để bạn duyệt bằng chứng thật trước P2. |
| Xoá trước, dọn sau | P0 xoá CLI + SQLite ngay để thấy rõ cái gì vỡ, rồi vá dần. Nhanh, không phải duy trì 2 đường song song. Nhưng repo mất khả năng tự vận hành giữa chừng — không còn check/handoff để gate chính việc đang làm. |
| Tách 2 release: v0.15 fail-open, v0.16 xoá CLI | v0.15 chỉ làm P0+P1 (fail-open + hook guard), CLI vẫn còn nhưng không ai bắt buộc dùng. Sống thật với nó một thời gian rồi mới xoá ở v0.16. An toàn nhất, nhưng kéo dài giai đoạn 2 control plane mà 0027 cảnh báo. |


---

## Spec accepted at the end of the interview

The interview closed with this spec, accepted by the owner before `to-plan` ran. It is
reproduced verbatim; its `Context to Read First` table is the direct ancestor of the merged
plan's Authority section, and its phase diagram is the ancestor of the five phases.


---
title: zharness v0.15 "slim" — Installer-only binary, markdown-only state, fail-open
status: validated
interviewed: 2026-08-26
supersedes: docs/plans/completed/zharness-v015-slim.md, .kit/plans/2026-08-26-zharness-v013-slim/plan.md
lane: high-risk
---

## Outcome

`zharness` chỉ còn là binary install/update/uninstall. Toàn bộ vòng đời
(brainstorm → to-plan → work → check → handoff) chạy được từ markdown +
git hook + script trong repo, **không cần binary có mặt**. SQLite và toàn
bộ lifecycle command bị xoá khỏi source. Hai guarantee fail-closed
(proof re-execution, independent judge cho lane high-risk) chuyển từ
trong binary sang pre-commit hook — hook đọc bytes đã staged, không tin
bất kỳ marker nào agent tự ghi.

## Success Condition

Người ngoài cuộc phỏng vấn verify được bằng 6 lệnh:

| # | Signal | Cách chứng minh |
|---|---|---|
| S1 | Kill-switch | `PATH` không có `zharness` → hoàn tất 1 task thật, ghi đúng `## Progress`, zero STOP |
| S2 | Proof guard | Commit thử 1 Validation entry `APPROVED` với proof command cố tình fail → hook reject. Sửa proof cho pass → hook cho qua |
| S3 | Judge guard | Commit entry `judge: same-session` vào plan có `lane: high-risk` → hook reject |
| S4 | Kill list sạch | `rg -i "sqlite\|harness\.db" cli/` = 0 (trừ note EOL trong CHANGELOG); `zharness --help` chỉ hiện install/update/uninstall |
| S5 | Identity test | Repo mới scaffold, session chưa mồi → trả lời đúng "là gì / kiến trúc sao / đang làm gì" chỉ từ `docs/PROJECT.md` + plans |
| S6 | Guard count | Đúng 2 fail-closed (cả hai ở pre-commit hook) + 1 pause-point (material product ambiguity, tầng playbook). Không còn cổng fail-closed nào khác |
| S7 | Không phình token | Cold-entry token vào `work` **không tăng quá 10%** so với 0.14.0. Đo bằng byte: `AGENTS.md + playbook + markdown agent phải đọc` so với `preflight packet (2.595 tok) + playbook`. Đếm byte, không cần CLI |

> S7 cũ ("−30% chain cost") **bị xoá**: audit `sdlc-token-cache-audit.md`
> dự báo −31% từ P1–P3, cả 3 đã ship (`work.md:41`, `work.md:45`,
> `plan_query_test.go:164`). Đòn bẩy model đã dùng hết; audit:202 nói
> phần v0.15 làm chỉ đáng ~$0.002/phase. Rủi ro thật là token **tăng**,
> nên đó mới là thứ đáng đo.

## Scope

**May change**
- `skills/workflow/{watzup,work,check,brainstorm,to-plan,handoff}/SKILL.md` — bỏ cả 2 tầng STOP
- `skills/workflow/README.md:39,55` — MIN_ZHARNESS_VERSION thôi làm cổng chặn
- `cli/docs/embedded/playbooks/*.md` (6 file, 65 call site) + `cli/docs/embedded/AGENTS.md`
- `cli/internal/**` — xoá lifecycle command + SQLite; giữ install/update/uninstall
- `scripts/record-check.sh` (mới), `scripts/install-git-hooks.sh` (mở rộng)
- `docs/PROJECT.md` (mới), `docs/ARCHITECTURE.md` (viết lại + repin), `docs/memory/` (mới)
- `.github/workflows/cli-ci.yml` — thêm job chạy lại hook guard

**Must not change**
- 2 cây skill **đã cài**: `~/.claude/skills`, `~/.agents/skills`, global rules, `~/.codex/AGENTS.md` — trừ đúng 1 dòng `project_doc_fallback_filenames = ["CLAUDE.md"]`
- `harness.db` của consumer và sidecar — consumer-owned bytes, không bao giờ xoá
- Hình dạng 6 stage pipeline — chỉ đổi cơ chế enforce
- `CLAUDE.md` giữ vai trò cầu `@AGENTS.md`
- Không backfill lịch sử giả

## Context to Read First

| File | Vì sao |
|---|---|
| `.kit/plans/2026-08-26-zharness-v013-slim/plan.md` | Authority gốc. **Đang bị `.gitignore:6` chặn** → phải copy decisions + research anchors + audit evidence vào plan gộp, nếu không CI và mọi clone đều mất |
| `cli/docs/CONTRACT.md:189` | Guarantee của `check record` mà hook phải giữ nguyên |
| `cli/docs/CONTRACT.md:50-56` | 5 việc `init` đang làm — installer nhận lại 4 |
| `docs/audit/sdlc-token-cache-audit.md:200-213` | Vì sao S7 cũ chết; audit tự bác việc gộp model |
| `docs/plans/completed/harness-fixes-63-64.md` (mục Validation) | Format thật của Validation entry — hook phải parse đúng |
| `docs/decisions/0004-docs-directory-deletion-655c6ac.md` | Tiền lệ mất plan record khi xoá docs |
| `AGENTS.md:4,8` | 2 câu fail-closed còn sót, S1 cần chúng biến mất |

## Key Decisions

1. **6 SKILL.md của repo bị sửa; NG1 thu hẹp** — NG1 cũ nói chúng "ngoài product path", sai: đây chính là thứ `npx skills add` cài. NG1 mới chỉ còn cấm đụng 2 cây **đã cài**. Mở khoá S1, S6.
2. **Guarantee nằm ở pre-commit hook, không ở marker** — hook đọc staged diff, tự trích proof command từ sub-bullet và **chạy lại**; grep `lane:` frontmatter + `judge:` entry để chặn high-risk + same-session. Không tin marker vì agent vừa gọi script vừa viết entry. `scripts/record-check.sh` chỉ là tiện ích chạy trước. Phụ thuộc: quyết định 1.
3. **S7 đổi từ "giảm 30%" sang "không tăng quá 10%"** — đòn bẩy cũ đã dùng hết, và method đo cũ (`init + intake + story + run create + trace add`) bị chính R1 xoá. Method mới đếm byte, sống sót qua R1.
4. **`install` = `init` trừ DB, cộng PROJECT.md** — `.zharness/base/` thay bảng `managed_docs` làm nơi track hash. Không thêm verb nào ngoài install/update/uninstall.
5. **`install` tự nhận brownfield, in report read-only** — deterministic detection (đếm active plans, dò README/CLAUDE.md/state-file lạ), in đề xuất, exit 0, không auto-rewrite. HARVEST/DIET/RECONCILE của v013 chuyển xuống playbook brainstorm. Phụ thuộc: quyết định 4.
6. **Kill list dựng lại từ `root.go`, không từ `CONTRACT.md`** — CONTRACT thiếu `plan complete`/`plan abandon` (`root.go:59`, `plan.go:16,21,30`); memory là ×4 không phải ×3 (`memory.go:22,40,51,73`); `status`/`doctor` **không tồn tại** ở 0.14.0 — đó là verb v013 định làm.
7. **`docs/ARCHITECTURE.md` viết lại ở phase cuối** — nó đang tả đúng kiến trúc sắp bị xoá ("SQLite là index", "binary là guardrail ghi markdown") và mang `zharness:pin`, nên audit sẽ báo pin drift nếu để nguyên.

## Validation Loop

**Trong lúc làm** — cổng mỗi phase, cả 4 phải xanh:
1. `bash scripts/verify-doc-links.sh`
2. `cd cli && go test ./...` — còn bao nhiêu chạy bấy nhiêu
3. `grep -rn` kill-list trong `docs/` + `skills/` → **0 hit** (bắt tên lệnh chết còn sót trong prose — `go test` không bắt được)
4. Kill-switch smoke: `PATH` không có binary → hoàn tất 1 task thật, `## Progress` đúng

**Chứng minh cuối** — S1–S7 ở trên, cộng:
- Hook guard chạy song song `check record` cũ ở P1, cả hai cho **cùng verdict** trên 1 case thật + 1 case cố tình sai
- Mỗi phase merge độc lập được

## Stop / Pause

**Thứ tự phase** (không đảo được):
