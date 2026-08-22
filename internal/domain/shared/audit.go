package shared

import "time"

type AuditEvent struct {
	ID         ID             `json:"id"`
	MallID     ID             `json:"mall_id"`
	ActorID    ID             `json:"actor_id"`
	Source     string         `json:"source"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID ID             `json:"resource_id"`
	Before     map[string]any `json:"before"`
	After      map[string]any `json:"after"`
	Reason     string         `json:"reason"`
	RequestID  string         `json:"request_id"`
	At         time.Time      `json:"at"`
}

func NewAuditEvent(id ID, actor Actor, action, resource string, resourceID ID, reason, requestID string, before, after map[string]any, at time.Time) AuditEvent {
	return AuditEvent{ID: id, MallID: actor.MallID, ActorID: actor.ID, Source: actor.Source, Action: action, Resource: resource, ResourceID: resourceID, Before: before, After: after, Reason: reason, RequestID: requestID, At: at.UTC()}
}
