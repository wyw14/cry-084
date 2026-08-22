package inventory

import (
	"fmt"

	"github.com/local/cry-084/internal/domain/shared"
)

type Part struct {
	ID       shared.ID
	MallID   shared.ID
	SKU      string
	Name     string
	Quantity int
	Reserved int
	Version  shared.Version
}

type Consumption struct {
	ID       shared.ID
	HazardID shared.ID
	PartID   shared.ID
	Quantity int
	ActorID  shared.ID
}

func (p *Part) Consume(id, hazardID, actorID shared.ID, quantity int) (Consumption, error) {
	if id == "" || hazardID == "" || actorID == "" || quantity < 1 {
		return Consumption{}, fmt.Errorf("%w: invalid consumption", shared.ErrValidation)
	}
	if p.Quantity-p.Reserved < quantity {
		return Consumption{}, fmt.Errorf("%w: insufficient stock", shared.ErrConflict)
	}
	p.Quantity -= quantity
	p.Version++
	return Consumption{ID: id, HazardID: hazardID, PartID: p.ID, Quantity: quantity, ActorID: actorID}, nil
}
