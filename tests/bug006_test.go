package tests

import (
	"testing"
	"github.com/local/cry-084/internal/domain/inventory"
)

func TestBug006ReservedPartsCannotBeConsumed(t *testing.T){p:=inventory.Part{ID:"p",Quantity:5,Reserved:4};if _,err:=p.Consume("c","h","u",2);err==nil{t.Fatal("reserved stock is unavailable")}}
