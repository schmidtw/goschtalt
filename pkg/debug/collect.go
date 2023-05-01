// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"bytes"
	"fmt"
	"sort"
)

// Collect is a simple implementation of a Reporter that collects all the mappings
// as a map and provides a clean way to output them.
type Collect struct {
	Mapping map[string]string
}

//var _ goschtalt.Reporter = (*Collect)(nil)

// Report stores the mapping information.
func (c *Collect) Report(a, b string) {
	if c.Mapping == nil {
		c.Reset()
	}
	c.Mapping[a] = b
}

// Reset clears out the mapping information.
func (c *Collect) Reset() {
	c.Mapping = make(map[string]string)
}

// String produces a sorted & pretty output string.
func (c Collect) String() string {
	keys := make([]string, 0, len(c.Mapping))
	max := 0
	for k := range c.Mapping {
		keys = append(keys, k)
		if len(k) > max {
			max = len(k)
		}
	}
	sort.Strings(keys)

	buf := new(bytes.Buffer)
	for _, k := range keys {
		fmt.Fprintf(buf, "'%s' % *s--> '%s'\n", k, (max - len(k)), "", c.Mapping[k])
	}

	return buf.String()
}
