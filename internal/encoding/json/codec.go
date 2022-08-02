// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"bytes"
	"encoding/json"
)

type Codec struct{}

func (c Codec) Decode(b []byte, v *map[string]any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(v)
	if err != nil {
		return err
	}

	fixMap(*v)

	return nil
}

func (c Codec) Encode(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

func (c Codec) Extensions() []string {
	return []string{"json"}
}

// The next few functions recursively walk the tree and change the numbers from
// strings (really json.Number) to either an int64 or float64 depending on if
// the number is an integer or not.  Numbers that are larger than an int64 will
// be converted to a float64.  Numbers beyond that range are left as a string.

func fixMap(m map[string]any) {
	for key, val := range m {
		m[key] = fixVal(val)
	}
}

func fixVal(val any) any {
	switch val := val.(type) {
	case map[string]any:
		fixMap(val)
	case []any:
		fixArray(val)
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return i
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	}
	return val
}

func fixArray(a []any) {
	for i, val := range a {
		a[i] = fixVal(val)
	}
}
