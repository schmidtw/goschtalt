// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"encoding/json"
)

type Codec struct{}

func (c Codec) Decode(b []byte, v *map[string]any) error {
	return json.Unmarshal(b, v)
}

func (c Codec) Encode(v *map[string]any) ([]byte, error) {
	return json.Marshal(v)
}

func (c Codec) Extensions() []string {
	return []string{"json"}
}
