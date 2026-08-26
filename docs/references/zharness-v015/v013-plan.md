# Archive: zharness v0.13 "slim" plan (superseded)

> **Archive — not a live plan.** Authored 2026-08-26 and superseded the same day by
> `docs/plans/active/zharness-v015-slim.md`. It lived at
> `.kit/plans/2026-08-26-zharness-v013-slim/plan.md`, which `.gitignore` excludes, so it
> reached no clone and no CI run. Copied here verbatim so the merged plan's Authority
> section can cite a real committed path instead of absorbing it inline.
>
> Every path, line number, and command name below is **as it was on 2026-08-26**. Do not
> treat them as live cross-references — several name commands that v0.15 deletes and files
> that do not exist yet.

---

# Plan: zharness v0.13 "slim" — Fail-open, Markdown-only

**Date:** 2026-08-26
**Baseline:** zharness 0.12.0 @ /Users/tinhtute/Lab/mono-harness
**Status:** approved, awaiting execution
**Prior art chính:** github.com/hoangnb24/repository-harness (đã đi cùng con đường — EOL SQLite protocol ngày 2026-08-10 theo decision 0027 của họ)

## Decisions locked (2026-08-26)

1. **Bỏ hẳn SQLite** — state duy nhất là committed markdown + git. Repository là system of record.
2. **Write command duy nhất: `record check`** (proof re-execution cơ khí). Mọi event khác (trace, decision, intake, story, run, handoff) agent tự append markdown theo playbook.
3. **Fail-closed chỉ còn 2 guard:** proof verification khi ghi APPROVED verdict + independent judge cho lane high-risk. Material product ambiguity = pause-point hợp pháp ở tầng playbook (không phải CLI error).
4. **v0.13 breaking change** — consumer cũ pin 0.12.x; không tự xoá `harness.db` của consumer (consumer-owned bytes, đúng bài học decision 0027); CHANGELOG breaking note.
5. **Gộp toàn bộ instruction layer toàn cục** (~/.claude + ~/.codex + 2 cây skills thành 1 nguồn).

## Research anchors

| Nguồn | Bài học áp dụng |
|---|---|
| hoangnb24/repository-harness decision 0027 | EOL playbook: pin bản cuối làm archive; một tree = một sản phẩm; không để `legacy/`; consumer bytes là consumer-owned |
| hoangnb24/repository-harness AGENTS.md | Entry point ~20 dòng zero-CLI-required; "repository remains the system of record"; work-shape routing; *"configurable defaults are not authority"*; *"no parallel control-plane state"* |
| hoangnb24/repository-harness updater | Three-way merge (.harness-core/ BASE/LOCAL/UPSTREAM), conflict dừng chờ human (`--continue/--abort`), activation transactional |
| hoangnb24/repository-harness decision 0020 | Knowledge boundaries: khai báo tường minh cài gì / không bao giờ cài gì — "no fabricated application truth"; `$onboard-repository` read-only-first, explicit-only |
| platform.claude.com prompt-caching | Prefix match tuyệt đối; cache model-scoped, model switch không có escape hatch; dynamic data (timestamp/session) phải nằm cuối hoặc loại bỏ |
| anthropic.com Agent Skills | Progressive disclosure: ~100 tokens metadata/skill luôn nạp; body <5k khi trigger; resources on-demand |
| systemdesignschool.io fail-open vs fail-closed | Fail open với capacity/ceremony guard; fail closed CHỈ với correctness/security guard; local fallback = fail-open có sàn |
| sujeet.pro graceful degradation | Degradation phải graduated (L0–L5), explicit, observable — không nhảy từ healthy xuống unavailable |
| agents.md (AAIF chuẩn) | 1 AGENTS.md canonical cho 30+ tools; CLAUDE.md chỉ cần `@AGENTS.md` import bridge; file ngắn viết tay hiệu quả hơn |
| github/spec-kit | Prior art SDLC thuần markdown + scripts tùy chọn, không DB, quy mô cộng đồng thật |

## Audit evidence (root causes đã xác minh trong repo này)

| Vấn đề | Bằng chứng |
|---|---|
| Fail-closed ×3 tầng ở mọi cổng vào | 6/6 SKILL.md hard-stop: `skills/workflow/{watzup/SKILL.md:14, work/SKILL.md:15, check/SKILL.md:16, brainstorm/SKILL.md:15, to-plan/SKILL.md:14, handoff/SKILL.md:14}` |
| DB unreadable chặn mọi stage | `cli/docs/CONTRACT.md:44` |
| invalid_stage/invalid_mode, CLI không tự mô tả | `cli/internal/domain/preflight.go:60-78` |
| AGENTS.md block stale (mâu thuẫn F3 đã fix ở skill nhưng block chưa) | `AGENTS.md:4` vs `skills/workflow/README.md:39` |
| Instruction trùng lặp 2 lớp global | `~/.codex/AGENTS.md` 4984B ≈ nội dung `~/.claude/rules/*` 7245B |
| 2 cây skills song song | `~/.claude/skills` 32 items (~296KB) + `~/.agents/skills` 29 items (~283KB) |
| DB chỉ là derived index — bỏ được | `cli/docs/CONTRACT.md:75` (db rebuild tái tạo toàn bộ lifecycle từ committed plan markdown một mình) |
| Memory hoàn chỉnh nhưng không ai gọi | `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` |
| Consumer CLAUDE.md phình, mô tả lại thứ filesystem tự thấy | `docs/audit/consumer-adoption-audit.md:204-211` (D4: 349 dòng / 3,169 tokens mỗi turn) |
| Packet `phases` unbounded | `docs/audit/consumer-adoption-audit.md` D2 |
| Token/cache: model-scoped cache + stage switching | `docs/audit/sdlc-token-cache-audit.md:18-28` ($0.275/phase chỉ cho model switch); `work.md:45` đã vá một phần |

## Target architecture

### State

State = git-committed markdown duy nhất:

```
docs/plans/{active,completed}/*.md   # lifecycle ledger (append-only Progress/
                                     # Decisions/Validation sections)
docs/PROJECT.md                      # identity ≤50 dòng (K0)
docs/memory/{id}.md                  # session memory convention (K4)
docs/decisions/*.md                  # ADR (K3)
```

### CLI surface — từ ~20 command groups xuống 3 verbs

1. **`zharness status --json`**
   Read aggregator duy nhất: position + phases (BOUNDED, có Omitted declaration — fix D2) + drift + PROJECT.md head (~10 dòng) + top-3 memories match keywords (cap ~300 tokens) + valid stages/modes (tự mô tả — diệt invalid_stage/invalid_mode).
   KHÔNG BAO GIỜ exit non-zero cho trạng thái thông tin. Đọc bằng slice plan markdown, không whole-file.

2. **`zharness record check '[payload]'`**
   Lệnh write DUY NHẤT — tồn tại vì enforce: re-execute mọi proof command trước khi ghi APPROVED (giữ hành vi CONTRACT.md:189); high-risk lane đọc từ plan frontmatter `lane:` (CONTRACT.md:108). Append `## Validation` entry vào active plan. File lock giữ lại chống race đa tiến trình.

3. **`zharness doctor`**
   Docs integrity + adopt-detection (brownfield findings) + refresh managed docs bằng THREE-WAY MERGE trên `.zharness/base/` (BASE/LOCAL/UPSTREAM giữ lại; conflict → dừng chờ human resolve, transactional). Thấy `harness.db` thừa → chỉ báo danh, KHÔNG xoá.

**Kill list:** preflight, resume, query×9 views, audit, init, migrate, import, db rebuild/status, intake, story, run create, trace add, decision add, handoff record, validate, intervention, id, memory×3, scaffold.
**Drop dependency:** modernc.org/sqlite + migrations + repository locks (SQLite).
trace/decision/intake/story/run/handoff: agent tự append đúng section plan theo playbook discipline ("CLI owns the pen" chỉ còn ở record check).

### Fail-open contract

Degradation graduated thay binary STOP:

```
L0 healthy        → CLI đầy đủ
L1 markdown-read  → mọi READ suy ra từ plan markdown khi DB/thành phần thiếu
L2 markdown-write → WRITE chỉ ghi markdown, ghi rõ trong response
L3 harness-absent → CLI chết/lỗi bất kỳ → đọc docs/WORKFLOW.md + playbook,
                    làm tiếp với git + plan files. KHÔNG STOP.
```

Pause-point hợp pháp duy nhất ngoài 2 guard: material product ambiguity ở tầng playbook ("add rate limiting" không có quota/key → stop-before-mutation, khuôn repository-harness WORKFLOW.md).

### Instruction layer

AGENTS.md block v2 theo khuôn repository-harness (~20 dòng):
- repository là system of record; work-shape routing (read-only / bounded / durable-planned)
- authority boundary + câu chốt "no task database, no parallel control-plane state"
- stage→playbook map chỉ là bảng accelerator TUỲ CHỌN
- zero CLI requirement để bắt đầu làm việc
- bỏ lệnh `zharness --version` riêng (F3 residue)

CLAUDE.md = `@AGENTS.md` import bridge (giữ nguyên — đã đúng).

## Consumer knowledge architecture

### Taxonomy — one question, one home

| Lớp | Câu hỏi | Nhà duy nhất | Sở hữu |
|---|---|---|---|
| K0 Identity | Là gì / cho ai / non-goals / chạy-test thế nào / đang giai đoạn nào | `docs/PROJECT.md` ≤50 dòng | Author, có bước ghi buộc |
| K1 Architecture | Build thế nào, vì sao | Docs sẵn có của consumer + pointers trong PROJECT.md | Author — harness KHÔNG nhân bản (bài học D4) |
| K2 Lifecycle | Đang làm tới đâu | `docs/plans/{active,completed}/*.md` | record check + agent appends |
| K3 Decisions | Vì sao đi đường này | ADR `docs/decisions/` + section append-only trong plan | Author + agent |
| K4 Memory | Học được gì giữa các session | `docs/memory/{id}.md` — agent grep trực tiếp | Agent, opt-in |
| K5 Instructions | Agent làm việc thế nào ở đây | Root `AGENTS.md` + ZHARNESS block + CLAUDE.md bridge; nested AGENTS.md cho monorepo | Mixed (block managed) |

Managed-doc marker nhúng thẳng vào file: `<!-- zharness:managed v0.13.0 sha=… -->`.

### Greenfield flow

1. `zharness init`/doctor scaffold-once: PROJECT.md template (5 câu hỏi chưa trả lời), docs/README.md map, decisions templates, `docs/memory/`.
2. **brainstorm lock = bước ghi buộc duy nhất**: trả lời 5 câu hỏi PROJECT.md lúc khoá SPEC (thay elicitation form cũ vốn chỉ dừng ở audit finding severity:info — CONTRACT.md:219 — tức không ai bao giờ phải trả lời).
3. to-plan/work/check/handoff ghi đúng phần knowledge nó sở hữu (kỷ luật "Owned Plan Sections" hiện có — work.md:19-28).

Knowledge mọc theo lifecycle, không làm ceremony documentation trước.

### Brownfield flow (read-only first)

1. `doctor --adopt` PHÁT HIỆN (deterministic): README/CLAUDE.md/AGENTS.md/docs/** tồn tại gì; bao nhiêu active plans; state-file lạ nào đang trả lời trùng câu hỏi.
2. Skill onboard (explicit-only, read-only trước):
   - HARVEST: draft PROJECT.md bằng cách trích xuất từ tài liệu sẵn có; user chỉ lấp khoảng trống
   - DIET: đề xuất CLAUDE.md chỉ giữ non-derivable gotchas/patterns (bài học D4) — ĐỀ XUẤT chờ duyệt, KHÔNG auto-rewrite
   - RECONCILE: ≥2 active plans → liệt kê bắt chọn; stale state-files → báo danh trash
3. Append ZHARNESS block, preserve toàn bộ nội dung hiện có.

Không backfill lịch sử giả (bài học issue #25).

### Knowledge boundaries declaration (khuôn 0020)

zharness cài: khung workflow + templates + pointer structure + maintenance binary.
zharness KHÔNG BAO GIỜ cài: product policy, application architecture, validation commands, credentials, database, schemas. Consumer nhận "no fabricated application truth".

## Phases

### P0 — Instruction fail-open [S]
Files: `cli/docs/embedded/AGENTS.md`, 6 SKILL.md, 6 embedded playbooks.
- Rewrite ZHARNESS block v2 (zero-CLI-required, work-shape routing)
- 6 SKILL.md: bỏ 2 tầng STOP (missing binary, version gate) → 1 dòng degradation
- Playbooks: purge ngôn ngữ STOP; CLI calls cũ → status/record/markdown-native
Verify: kill-switch test — PATH thiếu binary → work vẫn hoàn thành 1 task thật + ghi `## Progress`. Regenerate root AGENTS.md projection.

### P1 — Markdown-only core [M]
Viết status / record check / doctor (3-way merge); xoá kill-list; drop sqlite;
EOL playbook: CHANGELOG breaking note, consumer pin 0.12.x, doctor báo-list-không-xoá harness.db.
Verify:
- Toàn lifecycle chạy được qua 3 verbs HOẶC thuần markdown (binary absent)
- `rg -i "sqlite|harness.db" cli/` ≈ 0 (trừ EOL note)
- Doctor idempotent; crash giữa record check không corrupt markdown

### P2a/P2b — Knowledge baseline [S–M]
P2a greenfield: scaffold PROJECT.md + brainstorm-lock wiring + status packet fields.
P2b brownfield: onboard skill (read-only first) + memory-as-files convention.
Verify: fresh session, không mồi → trả lời đúng "dự án này là gì / kiến trúc ra sao / đang làm gì" chỉ từ repo.

### P3 — Global instruction merge [M]
1 nguồn rules dùng chung claude/codex (codex đọc qua `project_doc_fallback_filenames` trong config.toml); xoá SOUL/english trùng lặp; gộp 2 cây skills thành 1 cây + bridge.
Verify: `/context` mỗi client thấy đúng 1 bản; smoke-test 1 task trên claude-code + opencode + codex.

### P4 — Model policy & đo lại [S]
Bỏ `model:` pin khỏi watzup/to-plan/work/git/handoff (inherit session model);
opus chỉ brainstorm + check-full cuối initiative (kéo dài hướng đã ship ở work.md:45).
Chạy lại phương pháp đo `docs/audit/sdlc-token-cache-audit.md`.
Verify: chain cost −30% vs baseline 8/2026; cache-read ratio >80% single-session.

## Accepted risks (owner đã chấp nhận — dự án thử nghiệm)

- Mất compressed index cho plan rất dài → status slice theo section, chấp nhận chậm hơn ở quy mô lớn.
- Mất transaction đa bảng → ghi tuần tự 1 section + file lock; crash để lại markdown dở — doctor phát hiện, chấp nhận.
- Analytics đa initiative bằng grep thay vì SQL — đủ dùng ở quy mô này.
- Breaking change v0.13 → consumer cũ phải pin 0.12.x hoặc tự chuyển sang markdown thuần (plans đã chứa toàn bộ truth).

## Success criteria (toàn initiative)

1. Kill-switch: binary absent → task completes, zero STOP, markdown bookkeeping đúng.
2. Toàn bộ lifecycle chạy được qua status/record-check/doctor hoặc thuần markdown.
3. Fresh-session project identity test pass không mồi tay.
4. Chain cost −30%, cache-hit >80% single-session (đo theo method P4).
5. Đúng 2 guard fail-closed (proof verification, high-risk judge) + 1 pause-point playbook (material ambiguity) trong toàn hệ thống.
