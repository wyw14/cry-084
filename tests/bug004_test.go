package tests

import (
	"testing"
	"github.com/local/cry-084/internal/domain/route"
)

func TestBug004RouteRejectsDuplicateStops(t *testing.T){r:=route.Route{ID:"r-004",TeamID:"team",Stops:[]route.Stop{{AssetID:"a",Order:1,Minutes:2},{AssetID:"a",Order:2,Minutes:2}}};if _,err:=r.OrderedStops();err==nil{t.Fatal("a device cannot appear twice on one route")}}
