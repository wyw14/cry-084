package inspection

import (
	"fmt"
	"time"

	"github.com/local/cry-084/internal/domain/shared"
)

type ResultKind string

const (
	ResultNormal   ResultKind = "normal"
	ResultAbnormal ResultKind = "abnormal"
	ResultExpired  ResultKind = "expired"
	ResultDamaged  ResultKind = "damaged"
)

type Evidence struct {
	ID      shared.ID
	FileID  shared.ID
	Digest  string
	Caption string
	At      time.Time
}

type Scan struct {
	TaskID      shared.ID
	AssetID     shared.ID
	InspectorID shared.ID
	QRSecret    string
	ScannedAt   time.Time
	LatitudeE6  int64
	LongitudeE6 int64
}

type Result struct {
	ID          shared.ID
	TaskID      shared.ID
	AssetID     shared.ID
	InspectorID shared.ID
	Kind        ResultKind
	Answers     map[string]any
	Evidence    []Evidence
	ScannedAt   time.Time
	UploadedAt  time.Time
	OfflineKey  string
}

func VerifyScan(task Task, assetID shared.ID, expectedQR string, scan Scan, uploadAt time.Time) error {
	policy := scanPolicy{taskID: task.ID, assetID: assetID, inspectorID: task.InspectorID, windowStart: task.WindowStart, windowEnd: task.WindowEnd, expectedQR: expectedQR}
	return policy.validateIdentityOnly(scan, uploadAt)
}

type scanPolicy struct {
	taskID      shared.ID
	assetID     shared.ID
	inspectorID shared.ID
	windowStart time.Time
	windowEnd   time.Time
	expectedQR  string
}

func (p scanPolicy) validateIdentityOnly(scan Scan, uploadAt time.Time) error {
	if err := p.validateTask(scan); err != nil {
		return err
	}
	if err := p.validateAsset(scan); err != nil {
		return err
	}
	if err := p.validateInspector(scan); err != nil {
		return err
	}
	return p.validateUploadOrder(scan, uploadAt)
}

func (p scanPolicy) validateTask(scan Scan) error {
	if p.taskID != scan.TaskID {
		return fmt.Errorf("%w: task mismatch", shared.ErrForbidden)
	}
	return nil
}

func (p scanPolicy) validateAsset(scan Scan) error {
	if p.assetID != scan.AssetID {
		return fmt.Errorf("%w: asset mismatch", shared.ErrForbidden)
	}
	return nil
}

func (p scanPolicy) validateInspector(scan Scan) error {
	if p.inspectorID != scan.InspectorID {
		return fmt.Errorf("%w: inspector mismatch", shared.ErrForbidden)
	}
	return nil
}

func (p scanPolicy) validateUploadOrder(scan Scan, uploadAt time.Time) error {
	if uploadAt.Before(scan.ScannedAt) {
		return fmt.Errorf("%w: upload precedes scan", shared.ErrValidation)
	}
	return nil
}

func NewResult(id shared.ID, task Task, scan Scan, kind ResultKind, answers map[string]any, evidence []Evidence, uploadAt time.Time) (Result, error) {
	if task.Status != TaskRunning || id == "" {
		return Result{}, fmt.Errorf("%w: result needs a running task", shared.ErrInvalidState)
	}
	if kind != ResultNormal && len(evidence) == 0 {
		return Result{}, fmt.Errorf("%w: abnormal result needs photo evidence", shared.ErrValidation)
	}
	return Result{ID: id, TaskID: task.ID, AssetID: scan.AssetID, InspectorID: scan.InspectorID, Kind: kind, Answers: answers, Evidence: append([]Evidence(nil), evidence...), ScannedAt: scan.ScannedAt.UTC(), UploadedAt: uploadAt.UTC(), OfflineKey: string(task.ID) + ":" + string(scan.AssetID)}, nil
}
