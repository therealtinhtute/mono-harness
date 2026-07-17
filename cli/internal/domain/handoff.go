package domain

// HandoffAnchors is a Handoff's anchors JSON payload.
type HandoffAnchors struct {
	LatestRunID   *string  `json:"latest_run_id"`
	LatestCheckID *string  `json:"latest_check_id"`
	OpenItems     []string `json:"open_items"`
}

// Handoff has no producing command among CONTRACT.md's 19 (R6/R18 gap,
// see cli/docs/CONTRACT.md's escalation note). The table and struct exist
// now per SPEC R13 so no breaking migration is needed once the gap
// resolves.
type Handoff struct {
	ID        string
	RunID     *string
	CheckID   *string
	Anchors   HandoffAnchors
	CreatedAt string
}

func (h Handoff) Validate() error {
	return nil
}
