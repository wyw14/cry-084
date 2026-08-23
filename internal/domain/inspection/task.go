package inspection

import (
	"fmt"
	"time"

	"github.com/local/cry-084/internal/domain/shared"
	"github.com/local/cry-084/internal/domain/template"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskClaimed   TaskStatus = "claimed"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskExpired   TaskStatus = "expired"
	TaskCancelled TaskStatus = "cancelled"
)

type TaskAsset struct {
	AssetID shared.ID
	Order   int
	Checked bool
}

type Task struct {
	ID                shared.ID
	MallID            shared.ID
	RouteID           shared.ID
	ShiftID           shared.ID
	TeamID            shared.ID
	InspectorID       shared.ID
	WindowStart       time.Time
	WindowEnd         time.Time
	Status            TaskStatus
	Assets            []TaskAsset
	TemplateSnapshots map[shared.ID]template.Snapshot
	Version           shared.Version
}

func NewTask(id shared.ID, mallID, routeID, shiftID, teamID shared.ID, start, end time.Time, assets []TaskAsset, snapshots map[shared.ID]template.Snapshot) (Task, error) {
	if id == "" || mallID == "" || routeID == "" || shiftID == "" || teamID == "" || len(assets) == 0 || !end.After(start) {
		return Task{}, fmt.Errorf("%w: incomplete task", shared.ErrValidation)
	}
	return Task{ID: id, MallID: mallID, RouteID: routeID, ShiftID: shiftID, TeamID: teamID, WindowStart: start.UTC(), WindowEnd: end.UTC(), Status: TaskPending, Assets: append([]TaskAsset(nil), assets...), TemplateSnapshots: cloneSnapshots(snapshots), Version: 1}, nil
}

func cloneSnapshots(values map[shared.ID]template.Snapshot) map[shared.ID]template.Snapshot {
	result := make(map[shared.ID]template.Snapshot, len(values))
	for id, value := range values {
		value.Fields = append([]template.Field(nil), value.Fields...)
		result[id] = value
	}
	return result
}

func (t *Task) Claim(actor shared.Actor) error {
	if t.Status != TaskPending || actor.TeamID != t.TeamID || !actor.HasRole("inspector", "inspection_lead") {
		return fmt.Errorf("%w: task cannot be claimed", shared.ErrInvalidState)
	}
	t.InspectorID = actor.ID
	t.Status = TaskClaimed
	t.Version++
	return nil
}

func (t *Task) Begin(actor shared.Actor, at time.Time) error {
	if t.Status != TaskClaimed || t.InspectorID != actor.ID || at.Before(t.WindowStart) || !at.Before(t.WindowEnd) {
		return fmt.Errorf("%w: task is outside claim or time window", shared.ErrInvalidState)
	}
	t.Status = TaskRunning
	t.Version++
	return nil
}

func (t *Task) MarkChecked(assetID shared.ID) error {
	for index := range t.Assets {
		if t.Assets[index].AssetID == assetID {
			t.Assets[index].Checked = true
			t.Version++
			return nil
		}
	}
	return fmt.Errorf("%w: asset is not on the route", shared.ErrValidation)
}

func (t *Task) Complete() error {
	if t.Status != TaskRunning {
		return fmt.Errorf("%w: task is not running", shared.ErrInvalidState)
	}
	for _, item := range t.Assets {
		if !item.Checked {
			return fmt.Errorf("%w: all assets must be checked", shared.ErrInvalidState)
		}
	}
	t.Status = TaskCompleted
	t.Version++
	return nil
}
