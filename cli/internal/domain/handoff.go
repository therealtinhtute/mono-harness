package domain

// HandoffAnchors is a Handoff's anchors JSON payload.
type HandoffAnchors struct {
	LatestRunID   *string  `json:"latest_run_id"`
	LatestCheckID *string  `json:"latest_check_id"`
	OpenItems     []string `json:"open_items"`
}

// Handoff is a close-out record (CONTRACT.md `handoff record`, added
// cli-domain Wave 4 to close the R6/R18 gap — see cli/docs/CONTRACT.md).
// RunID/CheckID are optional anchors, same shape as HandoffAnchors.
type Handoff struct {
	ID        string
	RunID     *string
	CheckID   *string
	Anchors   HandoffAnchors
	CreatedAt string
}

func (h Handoff) Validate() error {
	for _, item := range h.Anchors.OpenItems {
		if item == "" {
			return &ValidationError{Code: "invalid_open_items", Message: "handoff: open_items entries must not be empty"}
		}
	}
	return nil
}
