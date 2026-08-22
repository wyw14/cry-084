package route

import (
	"fmt"
	"sort"
	"time"

	"github.com/local/cry-084/internal/domain/shared"
)

type Stop struct {
	AssetID shared.ID
	Order   int
	Minutes int
}

type Route struct {
	ID      shared.ID
	MallID  shared.ID
	Name    string
	TeamID  shared.ID
	Stops   []Stop
	Version shared.Version
}

func (r Route) OrderedStops() ([]Stop, error) {
	if r.ID == "" || r.TeamID == "" || len(r.Stops) == 0 {
		return nil, fmt.Errorf("%w: route identity, team and stops are required", shared.ErrValidation)
	}
	result := append([]Stop(nil), r.Stops...)
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	seenAssets := map[shared.ID]bool{}
	seenOrders := map[int]bool{}
	for index, stop := range result {
		if stop.AssetID == "" || stop.Order != index+1 || stop.Minutes < 1 || seenAssets[stop.AssetID] || seenOrders[stop.Order] {
			return nil, fmt.Errorf("%w: route stops must be unique and contiguous", shared.ErrValidation)
		}
		seenAssets[stop.AssetID] = true
		seenOrders[stop.Order] = true
	}
	return result, nil
}

type Shift struct {
	ID       shared.ID
	RouteID  shared.ID
	TeamID   shared.ID
	StartsAt time.Time
	EndsAt   time.Time
}

func (s Shift) Validate() error {
	if s.ID == "" || s.RouteID == "" || s.TeamID == "" || !s.EndsAt.After(s.StartsAt) {
		return fmt.Errorf("%w: invalid shift", shared.ErrValidation)
	}
	return nil
}

type Schedule struct {
	ID        shared.ID
	RouteID   shared.ID
	TeamID    shared.ID
	Timezone  string
	Weekdays  []time.Weekday
	StartHour int
	Window    time.Duration
	Active    bool
}

func (s Schedule) Next(from time.Time, location *time.Location) time.Time {
	local := from.In(location)
	allowed := map[time.Weekday]bool{}
	for _, weekday := range s.Weekdays {
		allowed[weekday] = true
	}
	for offset := 0; offset < 15; offset++ {
		day := local.AddDate(0, 0, offset)
		candidate := time.Date(day.Year(), day.Month(), day.Day(), s.StartHour, 0, 0, 0, location)
		if allowed[candidate.Weekday()] && candidate.After(local) {
			return candidate.UTC()
		}
	}
	return time.Time{}
}
