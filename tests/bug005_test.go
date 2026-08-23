package tests

import (
	"testing"
	"time"
	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/template"
)

func TestBug005TemplateSnapshotIsIndependent(t *testing.T){v:=template.Version{ID:"v",TemplateID:"t",AssetKind:asset.Extinguisher,Number:1,Fields:[]template.Field{{Key:"pressure",Label:"压力",Kind:template.Number}}};snap,err:=template.Capture(v,time.Now());if err!=nil{t.Fatal(err)};v.Fields[0].Label="被修改";if snap.Fields[0].Label!="压力"{t.Fatal("historical task snapshot must not change")}}
