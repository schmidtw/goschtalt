// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	tests := []struct {
		description string
		mapping     map[string]string
		expect      string
	}{
		{
			description: "basic test",
			mapping: map[string]string{
				"Blue":  "blue",
				"Hello": "hello",
				"Tuba":  "tuba",
			},
			expect: "'Blue'  --> 'blue'\n'Hello' --> 'hello'\n'Tuba'  --> 'tuba'\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			c := Collect{}

			for k, v := range tc.mapping {
				c.Report(k, v)
			}

			assert.Equal(tc.mapping, c.Mapping)
			assert.Equal(tc.expect, c.String())
		})
	}
}
