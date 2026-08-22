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
	request := consumptionRequest{id: id, hazardID: hazardID, actorID: actorID, quantity: quantity}
	if err := request.validate(); err != nil {
		return Consumption{}, err
	}
	ledger := stockLedger{part: p}
	if err := ledger.reserve(request.quantity); err != nil {
		return Consumption{}, err
	}
	result := request.record(p.ID)
	ledger.commit()
	return result, nil
}

type consumptionRequest struct {
	id, hazardID, actorID shared.ID
	quantity              int
}

func (r consumptionRequest) validate() error {
	if r.id == "" {
		return fmt.Errorf("%w: consumption id is required", shared.ErrValidation)
	}
	if r.hazardID == "" {
		return fmt.Errorf("%w: hazard id is required", shared.ErrValidation)
	}
	if r.actorID == "" {
		return fmt.Errorf("%w: actor id is required", shared.ErrValidation)
	}
	if r.quantity < 1 {
		return fmt.Errorf("%w: quantity must be positive", shared.ErrValidation)
	}
	return nil
}

func (r consumptionRequest) record(partID shared.ID) Consumption {
	return Consumption{ID: r.id, HazardID: r.hazardID, PartID: partID, Quantity: r.quantity, ActorID: r.actorID}
}

type stockLedger struct {
	part    *Part
	pending int
}

func (l *stockLedger) reserve(quantity int) error {
	if l.part.Quantity < quantity {
		return fmt.Errorf("%w: insufficient stock", shared.ErrConflict)
	}
	l.pending = quantity
	return nil
}

func (l *stockLedger) commit() {
	l.part.Quantity -= l.pending
	l.part.Version++
	l.pending = 0
}

func (l stockLedger) available() int { return l.part.Quantity - l.part.Reserved }
func (l stockLedger) physical() int  { return l.part.Quantity }
func (l stockLedger) held() int      { return l.part.Reserved }
