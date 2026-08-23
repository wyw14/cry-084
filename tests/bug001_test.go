package tests

import (
	"testing"
	"time"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
)

func TestBug001ScanRequiresWindowAndExactQR(t *testing.T) {
	task := inspection.Task{ID:"task-001", InspectorID:"u1", WindowStart:time.Unix(100,0), WindowEnd:time.Unix(200,0)}
	wrongQR := inspection.Scan{TaskID:"task-001",AssetID:"a1",InspectorID:"u1",QRSecret:"forged",ScannedAt:time.Unix(150,0)}
	if err:=inspection.VerifyScan(task,shared.ID("a1"),"real",wrongQR,time.Unix(151,0)); err==nil { t.Fatal("forged QR must be rejected") }
	outOfWindow:=inspection.Scan{TaskID:"task-001",AssetID:"a1",InspectorID:"u1",QRSecret:"real",ScannedAt:time.Unix(250,0)}
	if err:=inspection.VerifyScan(task,shared.ID("a1"),"real",outOfWindow,time.Unix(251,0)); err==nil { t.Fatal("out-of-window scan must be rejected") }
}
