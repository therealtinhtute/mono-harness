package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// TestRankingEval_Golden_HitAt3 is the golden retrieval eval (R4): 10 fixed
// queries over seeded Vietnamese memories, each must hit its target inside
// top 3. Includes mixed-diacritic pairs kiem tra<->kiểm tra, dong bo<->đồng bộ
// and one combining-mark decomposed query. Fails before fold tokenizer.
func TestRankingEval_Golden_HitAt3(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	// Seed 10 memories with Vietnamese diacritic bodies (precomposed).
	// Each body is deliberately distinct so ranking can discriminate.
	ids := make(map[string]string)
	seed := func(key, body string) {
		id, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", body)
		if err != nil {
			t.Fatalf("CreateMemory %s: %v", key, err)
		}
		ids[key] = id
	}
	seed("M1", "kiểm tra đồng bộ dữ liệu phân tán hệ thống lưu trữ")
	seed("M2", "xác thực người dùng phân quyền vai trò truy cập bảo mật")
	seed("M3", "quản lý phiên làm việc timeout không hoạt động xử lý")
	seed("M4", "đồng bộ hóa trạng thái kế hoạch pha làm việc harness database")
	seed("M5", "bảo mật khóa API token secret không lưu trong body")
	seed("M6", "rào cản quy trình kiểm thử tích hợp liên tục CI")
	seed("M7", "tài liệu hướng dẫn quy ước durable lesson ghi nhớ")
	seed("M8", "phục hồi dữ liệu markdown rebuild khôi phục index")
	seed("M9", "kiểm định chất lượng truy vấn xếp hạng relevance ranking")
	seed("M10", "cấu hình biến môi trường triển khai sản xuất hệ thống")

	// Combining-mark decomposed form for "kiểm" = k i e + 0302(circumflex)+0309(hook)+ m
	decomposedKiemTra := "ki" + "e\u0302\u0309" + "m tra"

	type queryCase struct {
		query  string
		target string // key in ids
		label  string
	}
	cases := []queryCase{
		{"kiem tra", "M1", "kiem tra (ascii) -> kiểm tra"},
		{"dong bo du lieu", "M1", "dong bo du lieu -> đồng bộ dữ liệu"},
		{"xac thuc nguoi dung", "M2", "xac thuc nguoi dung -> xác thực người dùng"},
		{"quan ly phien", "M3", "quan ly phien -> quản lý phiên"},
		{"dong bo hoa trang thai", "M4", "dong bo hoa trang thai -> đồng bộ hóa trạng thái"},
		{"bao mat khoa api", "M5", "bao mat khoa api -> bảo mật khóa API"},
		{"tai lieu huong dan", "M7", "tai lieu huong dan -> tài liệu hướng dẫn"},
		{"phuc hoi du lieu markdown", "M8", "phuc hoi du lieu markdown -> phục hồi dữ liệu markdown"},
		{"kiem dinh chat luong", "M9", "kiem dinh chat luong -> kiểm định chất lượng"},
		{"cau hinh bien moi truong", "M10", "cau hinh bien moi truong -> cấu hình biến môi trường"},
		{decomposedKiemTra, "M1", "decomposed kiểm tra -> kiểm tra"},
	}

	// We run 11 cases but require at least 10; the golden eval counts hit@3.
	passed := 0
	for _, tc := range cases {
		results, err := MemoryQueryRanked(db, tc.query, "", "", "")
		if err != nil {
			t.Fatalf("MemoryQueryRanked %q (%s): %v", tc.query, tc.label, err)
		}
		found := false
		top := 3
		if len(results) < top {
			top = len(results)
		}
		for i := 0; i < top; i++ {
			if results[i].ID == ids[tc.target] {
				found = true
				break
			}
		}
		if !found {
			topIDs := []string{}
			for i := 0; i < top && i < len(results); i++ {
				topIDs = append(topIDs, results[i].ID)
			}
			t.Errorf("hit@3 miss for query %q (%s): target %s (%s) not in top %d %v; all results %+v", tc.query, tc.label, tc.target, ids[tc.target], top, topIDs, results)
		} else {
			passed++
		}
	}
	// Require all cases to hit; if any missed, test fails above. This count proves 10/10.
	if passed != len(cases) {
		t.Fatalf("golden eval %d/%d hit@3, want %d/%d", passed, len(cases), len(cases), len(cases))
	}
}

// TestRankingEval_CaseFold ensures case-insensitive matching survives folding.
func TestRankingEval_CaseFold(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)
	id, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "Kiểm Tra Đồng Bộ Dữ Liệu")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	results, err := MemoryQueryRanked(db, "KIEM TRA DONG BO", "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	found := false
	for i := 0; i < 3 && i < len(results); i++ {
		if results[i].ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("case-fold hit@3 miss: KIEM TRA DONG BO should hit %s, got %+v", id, results)
	}
}

// TestDedupGate_SimilarShouldBeDetected proves R3 pre-insert dedup via ranking:
// a new summary that shares >=4 folded tokens with an existing entry is considered similar.
// Uses diacritic pair so that before fold the match is missed (score 0) and after fold score >=4.
func TestDedupGate_SimilarShouldBeDetected(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	// Existing entry with diacritics, many tokens.
	existingID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "kiểm tra đồng bộ dữ liệu hệ thống quản lý phiên bản chi tiết")
	if err != nil {
		t.Fatalf("CreateMemory existing: %v", err)
	}

	// Near-duplicate summary without diacritics, overlapping 6+ tokens.
	newSummary := "kiem tra dong bo du lieu he thong quan ly"
	results, err := MemoryQueryRanked(db, newSummary, "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked for dedup check: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("dedup ranking: no results for %q, want at least existing %s", newSummary, existingID)
	}
	best := results[0]
	if best.ID != existingID {
		t.Fatalf("dedup ranking best ID = %s, want existing %s, results %+v", best.ID, existingID, results)
	}
	if best.Score < 4 {
		t.Fatalf("dedup ranking best score = %d for %q, want >=4 (folded-token matches); results %+v", best.Score, newSummary, results)
	}
}

// TestDedupGate_NonSimilarNotFlagged ensures low-overlap entries are not flagged as similar.
func TestDedupGate_NonSimilarNotFlagged(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	if _, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "kiểm tra đồng bộ dữ liệu hệ thống"); err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}

	// Unrelated summary shares <2 tokens, should score <4.
	summary := "cấu hình biến môi trường khác biệt hoàn toàn"
	results, err := MemoryQueryRanked(db, summary, "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) > 0 && results[0].Score >= 4 {
		t.Fatalf("non-similar summary %q should not score >=4, got %+v", summary, results[0])
	}
}

// TestDedupGate_ForceBypassSucceeds simulates the --force override: even when a
// near-duplicate exists (score >=4), creation with force semantics should still succeed.
// Before dedup gate, CreateMemory always succeeds, so this test alone would pass;
// the failure signal comes from TestDedupGate_SimilarShouldBeDetected. We keep this
// test to prove that after the interfaces gate, --force still allows the write
// (application layer itself never refuses; the gate is in interfaces).
func TestDedupGate_ForceBypassSucceeds(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	existingID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "kiểm tra đồng bộ dữ liệu hệ thống quản lý phiên bản")
	if err != nil {
		t.Fatalf("CreateMemory existing: %v", err)
	}

	newSummary := "kiem tra dong bo du lieu he thong quan ly phiên bản mới"
	// Direct ranking should show similarity (score >=4) after fold.
	results, err := MemoryQueryRanked(db, newSummary, "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) == 0 || results[0].Score < 4 || results[0].ID != existingID {
		t.Fatalf("precondition: expected similar detection for force test, got %+v", results)
	}

	// Application CreateMemory itself should still allow the write (force semantics
	// live in interfaces, not in application). This proves the write path is not
	// permanently blocked.
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", newSummary)
	if err != nil {
		t.Fatalf("CreateMemory with force semantics (application layer) should succeed even when similar exists: %v", err)
	}
	if newID == "" || newID == existingID {
		t.Fatalf("CreateMemory returned invalid id %q", newID)
	}
}
