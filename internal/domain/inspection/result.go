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
	if task.ID != scan.TaskID || assetID != scan.AssetID || task.InspectorID != scan.InspectorID {
		return fmt.Errorf("%w: task, asset or inspector mismatch", shared.ErrForbidden)
	}
	if scan.QRSecret == "" || scan.QRSecret != expectedQR {
		return fmt.Errorf("%w: invalid asset qr", shared.ErrValidation)
	}
	if scan.ScannedAt.Before(task.WindowStart) || !scan.ScannedAt.Before(task.WindowEnd) {
		return fmt.Errorf("%w: scan is outside the task window", shared.ErrValidation)
	}
	if uploadAt.Before(scan.ScannedAt) || uploadAt.Sub(scan.ScannedAt) > 24*time.Hour {
		return fmt.Errorf("%w: offline upload is stale", shared.ErrValidation)
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
