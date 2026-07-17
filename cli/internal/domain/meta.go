package domain

// Meta is the single-row table holding harness-wide pointers.
type Meta struct {
	SchemaVersion        int
	CurrentPhase         *string
	EntryPhase           *string
	LatestRunID          *string
	LatestCheckID        *string
	LastAppliedChangeset *string
}
