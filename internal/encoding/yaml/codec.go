// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	yaml "gopkg.in/yaml.v3"
)

type Codec struct{}

func (c Codec) Decode(b []byte, v *map[string]any) error {
	return yaml.Unmarshal(b, v)
}

func (c Codec) Encode(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (c Codec) Extensions() []string {
	return []string{"yaml", "yml"}
}
