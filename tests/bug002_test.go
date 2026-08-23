package tests

import (
	"testing"
	"time"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/shared"
)

func TestBug002CriticalHazardNeedsIndependentReview(t *testing.T) {
	h,err:=hazard.New("h-002","mall","asset","result",hazard.PriorityCritical,time.Now().Add(time.Hour),[]shared.ID{"before"});if err!=nil{t.Fatal(err)}
	h.Status=hazard.AwaitingReview;h.AfterEvidence=[]shared.ID{"after"};when:=time.Now();h.RectifiedAt=&when
	if err:=h.CanClose();err==nil{t.Fatal("critical hazard without review must not close")}
}
