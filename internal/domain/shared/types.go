package shared

import (
	"errors"
	"fmt"
	"time"
)

type ID string

func (id ID) Validate(name string) error {
	if id == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

type Version uint64

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("version conflict")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidState   = errors.New("invalid state transition")
	ErrValidation     = errors.New("validation failed")
	ErrAlreadyHandled = errors.New("already handled")
)

type Actor struct {
	ID       ID
	MallID   ID
	TeamID   ID
	Roles    []string
	Source   string
	Timezone *time.Location
}

func (a Actor) HasRole(roles ...string) bool {
	for _, held := range a.Roles {
		for _, wanted := range roles {
			if held == wanted {
				return true
			}
		}
	}
	return false
}

type Page struct {
	Number int
	Size   int
	Sort   string
	Order  string
}

func (p Page) Normalize(allowedSort map[string]bool) (Page, error) {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = 20
	}
	if p.Size > 100 {
		p.Size = 100
	}
	if p.Sort == "" {
		p.Sort = "created_at"
	}
	if !allowedSort[p.Sort] {
		return Page{}, fmt.Errorf("%w: unsupported sort", ErrValidation)
	}
	if p.Order == "" {
		p.Order = "asc"
	}
	if p.Order != "asc" && p.Order != "desc" {
		return Page{}, fmt.Errorf("%w: unsupported order", ErrValidation)
	}
	return p, nil
}
