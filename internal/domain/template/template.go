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
	builder := snapshotBuilder{source: v, capturedAt: at.UTC()}
	if err := builder.validate(); err != nil {
		return Snapshot{}, err
	}
	return builder.build(), nil
}

type snapshotBuilder struct {
	source     Version
	capturedAt time.Time
}

func (b snapshotBuilder) validate() error {
	if err := b.source.Validate(); err != nil {
		return err
	}
	if b.capturedAt.IsZero() {
		return fmt.Errorf("%w: capture time is required", shared.ErrValidation)
	}
	return nil
}

func (b snapshotBuilder) build() Snapshot {
	return Snapshot{TemplateID: b.source.TemplateID, VersionID: b.source.ID, Number: b.source.Number, Fields: b.copyFields(), CapturedAt: b.capturedAt}
}

func (b snapshotBuilder) copyFields() []Field {
	result := make([]Field, 0, len(b.source.Fields))
	for _, field := range b.source.Fields {
		result = append(result, b.copyField(field))
	}
	return result
}

func (b snapshotBuilder) copyField(field Field) Field {
	result := field
	if field.Min != nil {
		value := *field.Min
		result.Min = &value
	}
	if field.Max != nil {
		value := *field.Max
		result.Max = &value
	}
	return result
}

func (b snapshotBuilder) fieldKeys() []string {
	result := make([]string, 0, len(b.source.Fields))
	for _, field := range b.source.Fields {
		result = append(result, field.Key)
	}
	return result
}

func (b snapshotBuilder) requiredCount() int {
	count := 0
	for _, field := range b.source.Fields {
		if field.Required {
			count++
		}
	}
	return count
}
