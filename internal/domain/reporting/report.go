package reporting

import (
	"sort"
	"time"

	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
)

type Coverage struct {
	MallID        shared.ID `json:"mall_id"`
	From          time.Time `json:"from"`
	To            time.Time `json:"to"`
	Expected      int       `json:"expected"`
	Checked       int       `json:"checked"`
	Abnormal      int       `json:"abnormal"`
	CoverageBasis int       `json:"coverage_basis_points"`
}

func CalculateCoverage(mallID shared.ID, from, to time.Time, expected int, results []inspection.Result) Coverage {
	seen := map[shared.ID]bool{}
	abnormal := 0
	for _, result := range results {
		if result.ScannedAt.Before(from) || !result.ScannedAt.Before(to) {
			continue
		}
		seen[result.AssetID] = true
		if result.Kind != inspection.ResultNormal {
			abnormal++
		}
	}
	basis := 0
	if expected > 0 {
		basis = len(seen) * 10000 / expected
	}
	return Coverage{MallID: mallID, From: from.UTC(), To: to.UTC(), Expected: expected, Checked: len(seen), Abnormal: abnormal, CoverageBasis: basis}
}

type TimelineItem struct {
	At       time.Time   `json:"at"`
	Kind     string      `json:"kind"`
	Summary  string      `json:"summary"`
	Evidence []shared.ID `json:"evidence"`
}

func Timeline(items []TimelineItem) []TimelineItem {
	result := append([]TimelineItem(nil), items...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].At.Before(result[j].At) })
	return result
}

func latestTimelineItem(items []TimelineItem) (TimelineItem, bool) {
	ordered := Timeline(items)
	if len(ordered) == 0 {
		return TimelineItem{}, false
	}
	return ordered[len(ordered)-1], true
}
