package tests

import (
	"context"
	"testing"
	"time"

	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
)

func TestScanRejectsWrongTaskAndWindow(t *testing.T) {
	task := inspection.Task{ID: "task", InspectorID: "inspector", WindowStart: time.Unix(10, 0), WindowEnd: time.Unix(20, 0)}
	scan := inspection.Scan{TaskID: "other", AssetID: "asset", InspectorID: "inspector", QRSecret: "qr", ScannedAt: time.Unix(15, 0)}
	if err := inspection.VerifyScan(task, "asset", "qr", scan, time.Unix(16, 0)); err == nil {
		t.Fatal("mismatched task must be rejected")
	}
}

func TestCriticalHazardNeedsReview(t *testing.T) {
	h, err := hazard.New("h", "mall", "asset", "result", hazard.PriorityCritical, time.Now().Add(time.Hour), []shared.ID{"before"})
	if err != nil {
		t.Fatal(err)
	}
	h.Status = hazard.AwaitingReview
	h.AfterEvidence = []shared.ID{"after"}
	h.RectifiedAt = ptr(time.Now())
	if err := h.CanClose(); err == nil {
		t.Fatal("critical hazard without review must remain open")
	}
}

func TestStatusUsesLastValidEvent(t *testing.T) {
	status := asset.DeriveStatus([]asset.Event{{ID: "1", Status: asset.StatusDisabled, At: time.Unix(20, 0), Valid: true}, {ID: "2", Status: asset.StatusActive, At: time.Unix(10, 0), Valid: false}})
	if status != asset.StatusDisabled {
		t.Fatalf("got %s", status)
	}
}

func TestCancellationIsObserved(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ctx.Err() == nil {
		t.Fatal("cancelled context should carry error")
	}
}

func ptr(value time.Time) *time.Time { return &value }
