package request

import (
	"bytes"
	"encoding/json"
)

type OptionalPatch[T any] struct {
	Present bool
	Value   *T
}

func (o *OptionalPatch[T]) UnmarshalJSON(b []byte) error {
	o.Present = true
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		o.Value = nil
		return nil
	}

	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	o.Value = &v
	return nil
}
