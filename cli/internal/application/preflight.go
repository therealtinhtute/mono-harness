package application

import (
	"fmt"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

const (
	PreflightDBReady      = "ready"
	PreflightDBMissing    = "missing"
	PreflightDBUnreadable = "unreadable"

	PreflightDocsReady   = "ready"
	PreflightDocsMissing = "missing"
	PreflightDocsStale   = "stale"

	PreflightReady   = "ready"
	PreflightReduced = "reduced"
	PreflightBlocked = "blocked"
)

type PreflightView struct {
	Stage     string    `json:"stage"`
	Mode      string    `json:"mode"`
	DB        string    `json:"db"`
	Docs      string    `json:"docs"`
	Playbook  string    `json:"playbook,omitempty"`
	Readiness string    `json:"readiness"`
	Stop      *StopInfo `json:"stop,omitempty"`
}

func Preflight(stage, requestedMode, dbStatus, docsStatus, playbook string) (PreflightView, error) {
	stage = strings.ToLower(strings.TrimSpace(stage))
	mode, err := domain.ResolvePreflightMode(stage, requestedMode)
	if err != nil {
		return PreflightView{}, err
	}
	if !validPreflightDBStatus(dbStatus) {
		return PreflightView{}, fmt.Errorf("preflight: invalid db status %q", dbStatus)
	}
	if !validPreflightDocsStatus(docsStatus) {
		return PreflightView{}, fmt.Errorf("preflight: invalid docs status %q", docsStatus)
	}

	view := PreflightView{
		Stage:     stage,
		Mode:      mode,
		DB:        dbStatus,
		Docs:      docsStatus,
		Playbook:  playbook,
		Readiness: PreflightReady,
	}

	if dbStatus == PreflightDBUnreadable {
		view.Readiness = PreflightBlocked
		view.Stop = &StopInfo{
			Code:     "db_unreadable",
			Message:  "The harness database exists but cannot be read safely.",
			Recovery: "restore or rebuild harness.db from .kit/changesets before continuing",
		}
		return view, nil
	}

	if mode == domain.PreflightModeDurable {
		switch {
		case dbStatus == PreflightDBMissing:
			view.Readiness = PreflightBlocked
			view.Stop = &StopInfo{
				Code:     "harness_required",
				Message:  "This durable workflow stage requires an initialized harness database.",
				Recovery: "zharness init",
			}
		case docsStatus == PreflightDocsMissing:
			view.Readiness = PreflightBlocked
			view.Stop = &StopInfo{
				Code:     "docs_missing",
				Message:  "This durable workflow stage requires the managed workflow docs.",
				Recovery: "zharness init",
			}
		case docsStatus == PreflightDocsStale:
			view.Readiness = PreflightBlocked
			view.Stop = &StopInfo{
				Code:     "stale_docs",
				Message:  "The managed workflow docs do not match the running zharness version.",
				Recovery: "zharness init --refresh-docs",
			}
		}
		return view, nil
	}

	if dbStatus != PreflightDBReady || docsStatus != PreflightDocsReady {
		view.Readiness = PreflightReduced
	}
	return view, nil
}

func validPreflightDBStatus(status string) bool {
	switch status {
	case PreflightDBReady, PreflightDBMissing, PreflightDBUnreadable:
		return true
	default:
		return false
	}
}

func validPreflightDocsStatus(status string) bool {
	switch status {
	case PreflightDocsReady, PreflightDocsMissing, PreflightDocsStale:
		return true
	default:
		return false
	}
}
