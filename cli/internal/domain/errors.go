package domain

// ValidationError is a rejected entity's failure. Code is CONTRACT.md's
// stable machine-readable string when the rule maps to a documented one
// (e.g. "invalid_lane"), or "" for an internal invariant that CONTRACT.md
// does not enumerate (not reachable via a documented command argument).
// Domain returns this instead of importing interfaces' cliError, keeping
// the package free of other internal-package imports (layering rule);
// interfaces maps Code 1:1 into the JSON error envelope.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
