package template

import (
	"fmt"
	"time"

	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/shared"
)

type FieldKind string

const (
	Boolean FieldKind = "boolean"
	Number  FieldKind = "number"
	Text    FieldKind = "text"
	Photo   FieldKind = "photo"
)

type Field struct {
	Key      string
	Label    string
	Kind     FieldKind
	Required bool
	Min      *int
	Max      *int
}

type Version struct {
	ID          shared.ID
	TemplateID  shared.ID
	AssetKind   asset.Kind
	Number      int
	Fields      []Field
	PublishedAt time.Time
}

func (v Version) Validate() error {
	if v.ID == "" || v.TemplateID == "" || v.Number < 1 || len(v.Fields) == 0 {
		return fmt.Errorf("%w: incomplete template version", shared.ErrValidation)
	}
	seen := map[string]bool{}
	for _, field := range v.Fields {
		if field.Key == "" || field.Label == "" || seen[field.Key] {
			return fmt.Errorf("%w: invalid or duplicate template field", shared.ErrValidation)
		}
		seen[field.Key] = true
	}
	return nil
}

type Snapshot struct {
	TemplateID shared.ID
	VersionID  shared.ID
	Number     int
	Fields     []Field
	CapturedAt time.Time
}

func Capture(v Version, at time.Time) (Snapshot, error) {
	if err := v.Validate(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{TemplateID: v.TemplateID, VersionID: v.ID, Number: v.Number, Fields: append([]Field(nil), v.Fields...), CapturedAt: at.UTC()}, nil
}
