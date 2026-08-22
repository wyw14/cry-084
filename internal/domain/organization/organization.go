package organization

import (
	"fmt"

	"github.com/local/cry-084/internal/domain/shared"
)

type Mall struct {
	ID      shared.ID
	Name    string
	Address string
	Version shared.Version
}

type Floor struct {
	ID     shared.ID
	MallID shared.ID
	Name   string
	Level  int
}

type Zone struct {
	ID      shared.ID
	FloorID shared.ID
	Name    string
	Risk    RiskLevel
}

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type TeamKind string

const (
	InspectionTeam TeamKind = "inspection"
	RepairTeam     TeamKind = "repair"
)

type Team struct {
	ID     shared.ID
	MallID shared.ID
	Name   string
	Kind   TeamKind
}

type Membership struct {
	UserID shared.ID
	TeamID shared.ID
	Roles  []string
}

func Authorize(actor shared.Actor, mallID shared.ID, roles ...string) error {
	if actor.MallID != mallID {
		return fmt.Errorf("%w: resource belongs to another mall", shared.ErrForbidden)
	}
	if !actor.HasRole(roles...) {
		return fmt.Errorf("%w: role is not allowed", shared.ErrForbidden)
	}
	return nil
}
