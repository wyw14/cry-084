package asset

import (
	"fmt"
	"sort"
	"time"

	"github.com/local/cry-084/internal/domain/shared"
)

type Kind string

const (
	Hydrant        Kind = "hydrant"
	Extinguisher   Kind = "extinguisher"
	SmokeSensor    Kind = "smoke_sensor"
	Sprinkler      Kind = "sprinkler"
	EmergencyLight Kind = "emergency_light"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusWarning   Status = "warning"
	StatusDisabled  Status = "disabled"
	StatusRepairing Status = "repairing"
	StatusRetired   Status = "retired"
)

type Asset struct {
	ID             shared.ID
	MallID         shared.ID
	ZoneID         shared.ID
	Code           string
	QRSecret       string
	Kind           Kind
	Manufacturer   string
	Model          string
	InstalledAt    time.Time
	ExpiresAt      time.Time
	PressureUnit   string
	PressureMin    int
	CleanEveryDays int
	Status         Status
	Version        shared.Version
}

func (a Asset) Validate() error {
	if err := a.ID.Validate("asset id"); err != nil {
		return err
	}
	if a.Code == "" || a.QRSecret == "" || a.ZoneID == "" || a.MallID == "" {
		return fmt.Errorf("%w: asset identity and location are required", shared.ErrValidation)
	}
	switch a.Kind {
	case Hydrant, Extinguisher, SmokeSensor, Sprinkler, EmergencyLight:
	default:
		return fmt.Errorf("%w: unsupported asset kind", shared.ErrValidation)
	}
	return nil
}

type EventType string

const (
	EventRegistered    EventType = "registered"
	EventInspected     EventType = "inspected"
	EventDisabled      EventType = "disabled"
	EventRepairStarted EventType = "repair_started"
	EventRestored      EventType = "restored"
	EventRetired       EventType = "retired"
)

type Event struct {
	ID          shared.ID
	AssetID     shared.ID
	Type        EventType
	Status      Status
	Reason      string
	At          time.Time
	Valid       bool
	EvidenceIDs []shared.ID
}

func DeriveStatus(events []Event) Status {
	valid := make([]Event, 0, len(events))
	for _, event := range events {
		if event.Valid {
			valid = append(valid, event)
		}
	}
	if len(valid) == 0 {
		return StatusActive
	}
	sort.SliceStable(valid, func(i, j int) bool {
		if valid[i].At.Equal(valid[j].At) {
			return valid[i].ID < valid[j].ID
		}
		return valid[i].At.Before(valid[j].At)
	})
	return valid[len(valid)-1].Status
}

func RestoreEvent(id shared.ID, asset Asset, reason string, evidence []shared.ID, at time.Time) (Event, error) {
	if asset.Status != StatusDisabled && asset.Status != StatusRepairing && asset.Status != StatusWarning {
		return Event{}, fmt.Errorf("%w: asset is not recoverable", shared.ErrInvalidState)
	}
	if reason == "" || len(evidence) < 2 {
		return Event{}, fmt.Errorf("%w: restoration needs reason and before/after evidence", shared.ErrValidation)
	}
	return Event{ID: id, AssetID: asset.ID, Type: EventRestored, Status: StatusActive, Reason: reason, At: at.UTC(), Valid: true, EvidenceIDs: append([]shared.ID(nil), evidence...)}, nil
}
