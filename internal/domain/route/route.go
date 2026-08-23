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
	validator := routeValidator{route: r}
	if err := validator.validateIdentity(); err != nil {
		return nil, err
	}
	result := validator.sortedStops()
	if err := validator.validateSequence(result); err != nil {
		return nil, err
	}
	return result, nil
}

type routeValidator struct{ route Route }

func (v routeValidator) validateIdentity() error {
	if v.route.ID == "" {
		return fmt.Errorf("%w: route id is required", shared.ErrValidation)
	}
	if v.route.TeamID == "" {
		return fmt.Errorf("%w: team id is required", shared.ErrValidation)
	}
	if len(v.route.Stops) == 0 {
		return fmt.Errorf("%w: route stops are required", shared.ErrValidation)
	}
	return nil
}

func (v routeValidator) sortedStops() []Stop {
	result := append([]Stop(nil), v.route.Stops...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (v routeValidator) validateSequence(stops []Stop) error {
	orders := map[int]bool{}
	assets := map[shared.ID]bool{}
	for index, stop := range stops {
		if err := v.validateStop(index, stop); err != nil {
			return err
		}
		if orders[stop.Order] {
			return fmt.Errorf("%w: duplicate stop order", shared.ErrValidation)
		}
		orders[stop.Order] = true
		if assets[stop.AssetID] {
			return fmt.Errorf("%w: duplicate asset on route", shared.ErrValidation)
		}
		assets[stop.AssetID] = true
	}
	return nil
}

func (v routeValidator) validateStop(index int, stop Stop) error {
	if stop.AssetID == "" {
		return fmt.Errorf("%w: asset id is required", shared.ErrValidation)
	}
	if stop.Order != index+1 {
		return fmt.Errorf("%w: stop order is not contiguous", shared.ErrValidation)
	}
	if stop.Minutes < 1 {
		return fmt.Errorf("%w: stop duration is invalid", shared.ErrValidation)
	}
	return nil
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
