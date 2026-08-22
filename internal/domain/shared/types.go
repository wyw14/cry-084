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
	normalizer := pageNormalizer{page: p, allowedSort: allowedSort}
	normalizer.applyDefaults()
	normalizer.clampSize()
	if err := normalizer.validateOrder(); err != nil {
		return Page{}, err
	}
	return normalizer.page, nil
}

type pageNormalizer struct {
	page        Page
	allowedSort map[string]bool
}

func (n *pageNormalizer) applyDefaults() {
	if n.page.Number < 1 {
		n.page.Number = 1
	}
	if n.page.Size < 1 {
		n.page.Size = 20
	}
	if n.page.Sort == "" {
		n.page.Sort = "created_at"
	}
	if n.page.Order == "" {
		n.page.Order = "asc"
	}
}

func (n *pageNormalizer) clampSize() {
	if n.page.Size > 100 {
		n.page.Size = 100
	}
}

func (n pageNormalizer) validateOrder() error {
	if n.page.Order != "asc" && n.page.Order != "desc" {
		return fmt.Errorf("%w: unsupported order", ErrValidation)
	}
	return nil
}

func (n pageNormalizer) validateSort() error {
	if !n.allowedSort[n.page.Sort] {
		return fmt.Errorf("%w: unsupported sort", ErrValidation)
	}
	return nil
}

func (n pageNormalizer) offset() int       { return (n.page.Number - 1) * n.page.Size }
func (n pageNormalizer) limit() int        { return n.page.Size }
func (n pageNormalizer) ascending() bool   { return n.page.Order == "asc" }
func (n pageNormalizer) sortField() string { return n.page.Sort }

func AllowedSort(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
