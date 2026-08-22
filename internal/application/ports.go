package application

import (
	"context"
	"io"
	"time"

	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/inventory"
	"github.com/local/cry-084/internal/domain/route"
	"github.com/local/cry-084/internal/domain/shared"
	"github.com/local/cry-084/internal/domain/template"
)

type Repository interface {
	WithinTx(context.Context, func(Repository) error) error
	Asset(context.Context, shared.ID) (asset.Asset, error)
	SaveAsset(context.Context, asset.Asset, shared.Version) error
	AssetEvents(context.Context, shared.ID) ([]asset.Event, error)
	AppendAssetEvent(context.Context, asset.Event) error
	Route(context.Context, shared.ID) (route.Route, error)
	Schedule(context.Context, shared.ID) (route.Schedule, error)
	DueSchedules(context.Context, time.Time) ([]route.Schedule, error)
	TemplateVersion(context.Context, shared.ID) (template.Version, error)
	Task(context.Context, shared.ID) (inspection.Task, error)
	SaveTask(context.Context, inspection.Task, shared.Version) error
	CreateTask(context.Context, inspection.Task) error
	ResultByOfflineKey(context.Context, string) (inspection.Result, error)
	CreateResult(context.Context, inspection.Result) error
	Hazard(context.Context, shared.ID) (hazard.Hazard, error)
	SaveHazard(context.Context, hazard.Hazard, shared.Version) error
	CreateHazard(context.Context, hazard.Hazard) error
	Part(context.Context, shared.ID) (inventory.Part, error)
	SavePart(context.Context, inventory.Part, shared.Version) error
	CreateConsumption(context.Context, inventory.Consumption) error
	AppendAudit(context.Context, shared.AuditEvent) error
}

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() shared.ID }

type FileMetadata struct {
	ID        shared.ID
	MallID    shared.ID
	Name      string
	MIME      string
	Size      int64
	Digest    string
	CreatedAt time.Time
}

func (m FileMetadata) IsEvidence() bool { return m.MIME == "image/jpeg" || m.MIME == "image/png" }
func (m FileMetadata) IsDocument() bool { return m.MIME == "application/pdf" }

type FileStore interface {
	Put(context.Context, FileMetadata, io.Reader) error
	Open(context.Context, shared.ID) (FileMetadata, io.ReadCloser, error)
}

type Notification struct {
	ID      shared.ID
	MallID  shared.ID
	UserID  shared.ID
	Topic   string
	Message string
	DueAt   time.Time
}

type Notifier interface {
	Send(context.Context, Notification) error
}

type OutboxMessage struct {
	ID          shared.ID
	Topic       string
	Payload     []byte
	Attempts    int
	AvailableAt time.Time
	LastError   string
}

type Outbox interface {
	Enqueue(context.Context, OutboxMessage) error
	Claim(context.Context, int, time.Time) ([]OutboxMessage, error)
	Complete(context.Context, shared.ID) error
	Retry(context.Context, shared.ID, string, time.Time) error
	DeadLetter(context.Context, shared.ID, string) error
}
