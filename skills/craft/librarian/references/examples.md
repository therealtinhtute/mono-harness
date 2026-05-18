---
title: librarian — example research sessions
description: Sample GitHub research outputs from real queries
status: active
tags: [librarian, examples, github]
---

### Example 1 — Find where a symbol is defined in a repo

**Query:** "Where is `RunE` set on cobra commands in cli/cli?"

**Steps:**
1. `gh search code "RunE" --repo cli/cli --limit 10`
2. Cache top result: `gh api repos/cli/cli/contents/pkg/cmd/root.go?ref=main`
3. `rg -n "RunE" .kit/cache/github/cli/cli/pkg/cmd/root.go`

**Findings:**
- `.kit/cache/github/cli/cli/pkg/cmd/root.go:88` — `RunE` assigned inline as a closure
- Pattern: every subcommand sets `RunE`, never `Run`, for consistent error propagation

---

### Example 2 — Explore repo structure before deep diving

**Query:** "What's in the supabase-js SDK, especially auth?"

**Steps:**
1. `gh repo view supabase/supabase-js --json description,homepageUrl`
2. `gh api repos/supabase/supabase-js/git/trees/main?recursive=1 --jq '.tree[] | select(.path | startswith("src")) | .path'`
3. Cache `src/SupabaseClient.ts` and `src/lib/fetch.ts`

**Findings:**
- Auth lives in `src/lib/SupabaseAuthClient.ts` — wraps `@supabase/auth-js`
- Fetch is centralized in `src/lib/fetch.ts:12-34` — all requests go through `_fetch`

---

### Example 3 — Research API usage patterns across repos

**Query:** "How do popular CLIs handle config precedence (flags > env > file)?"

**Steps:**
1. `gh search code "viper.BindPFlag" --language go --limit 20`
2. Cache `cobra/cobra` config examples and `spf13/viper` README
3. `gh search code "os.Getenv" --repo charmbracelet/bubbletea --limit 5`

**Findings:**
- `spf13/viper` at `viper.go:312-340` — canonical layered config: flag > env > config > default
- `charmbracelet/bubbletea` uses env only for debug flags, not user config
- Recommendation: use viper with `AutomaticEnv()` + `BindPFlags()`

**Saved to:** `.kit/reports/github/cli-config-patterns.md`
