package hazard

import (
	"fmt"
	"time"

	"github.com/local/cry-084/internal/domain/shared"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type Status string

const (
	Open           Status = "open"
	Assigned       Status = "assigned"
	Arrived        Status = "arrived"
	Rectifying     Status = "rectifying"
	AwaitingReview Status = "awaiting_review"
	Closed         Status = "closed"
	Cancelled      Status = "cancelled"
)

type Hazard struct {
	ID             shared.ID
	MallID         shared.ID
	AssetID        shared.ID
	ResultID       shared.ID
	Priority       Priority
	Status         Status
	OwnerTeamID    shared.ID
	AssigneeID     shared.ID
	DueAt          time.Time
	ArrivedAt      *time.Time
	RectifiedAt    *time.Time
	ReviewedAt     *time.Time
	BeforeEvidence []shared.ID
	AfterEvidence  []shared.ID
	Reason         string
	Escalation     int
	Version        shared.Version
}

func New(id, mallID, assetID, resultID shared.ID, priority Priority, dueAt time.Time, before []shared.ID) (Hazard, error) {
	if id == "" || mallID == "" || assetID == "" || resultID == "" || dueAt.IsZero() || len(before) == 0 {
		return Hazard{}, fmt.Errorf("%w: incomplete hazard", shared.ErrValidation)
	}
	return Hazard{ID: id, MallID: mallID, AssetID: assetID, ResultID: resultID, Priority: priority, Status: Open, DueAt: dueAt.UTC(), BeforeEvidence: append([]shared.ID(nil), before...), Version: 1}, nil
}

func (h *Hazard) Assign(teamID, assigneeID shared.ID) error {
	if h.Status != Open && h.Status != Assigned {
		return fmt.Errorf("%w: hazard cannot be assigned", shared.ErrInvalidState)
	}
	if teamID == "" || assigneeID == "" {
		return fmt.Errorf("%w: repair team and assignee are required", shared.ErrValidation)
	}
	h.OwnerTeamID, h.AssigneeID, h.Status = teamID, assigneeID, Assigned
	h.Version++
	return nil
}

func (h *Hazard) MarkArrival(actor shared.Actor, at time.Time) error {
	if h.Status != Assigned || actor.ID != h.AssigneeID {
		return fmt.Errorf("%w: only assigned repairer can arrive", shared.ErrInvalidState)
	}
	at = at.UTC()
	h.ArrivedAt, h.Status = &at, Arrived
	h.Version++
	return nil
}

func (h *Hazard) SubmitRectification(actor shared.Actor, evidence []shared.ID, reason string, at time.Time) error {
	if h.Status != Arrived && h.Status != Rectifying || actor.ID != h.AssigneeID {
		return fmt.Errorf("%w: rectification is unavailable", shared.ErrInvalidState)
	}
	if len(evidence) == 0 || reason == "" {
		return fmt.Errorf("%w: after evidence and reason are required", shared.ErrValidation)
	}
	at = at.UTC()
	h.AfterEvidence = append([]shared.ID(nil), evidence...)
	h.Reason, h.RectifiedAt, h.Status = reason, &at, AwaitingReview
	h.Version++
	return nil
}

func (h *Hazard) Review(actor shared.Actor, accepted bool, at time.Time) error {
	if h.Status != AwaitingReview || !actor.HasRole("inspection_lead", "safety_manager") {
		return fmt.Errorf("%w: hazard is not awaiting an authorized review", shared.ErrInvalidState)
	}
	if !accepted {
		h.Status = Rectifying
		h.Version++
		return nil
	}
	at = at.UTC()
	h.ReviewedAt, h.Status = &at, Closed
	h.Version++
	return nil
}

func (h Hazard) CanClose() error {
	if h.Priority == PriorityCritical && h.ReviewedAt == nil {
		return fmt.Errorf("%w: critical hazard requires review", shared.ErrInvalidState)
	}
	if h.RectifiedAt == nil || len(h.BeforeEvidence) == 0 || len(h.AfterEvidence) == 0 {
		return fmt.Errorf("%w: rectification evidence is incomplete", shared.ErrInvalidState)
	}
	return nil
}

func (h *Hazard) Escalate(now time.Time) bool {
	if h.Status == Closed || h.Status == Cancelled || !now.After(h.DueAt) {
		return false
	}
	h.Escalation++
	h.Version++
	return true
}
