package tests

import (
	"testing"
	"time"
	"github.com/local/cry-084/internal/domain/asset"
)

func TestBug003StatusIgnoresInvalidLaterEvent(t *testing.T) {
	events:=[]asset.Event{{ID:"valid",Status:asset.StatusActive,At:time.Unix(100,0),Valid:true},{ID:"invalid",Status:asset.StatusDisabled,At:time.Unix(200,0),Valid:false}}
	if got:=asset.DeriveStatus(events);got!=asset.StatusActive{t.Fatalf("expected active from last valid event, got %s",got)}
}
