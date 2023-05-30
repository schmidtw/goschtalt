// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExplainerRecord(t *testing.T) {
	tests := []struct {
		in   ExplanationRecord
		want string
	}{
		{
			in:   ExplanationRecord{},
			want: "",
		}, {
			in: ExplanationRecord{
				Name:     "non-default-record",
				Duration: time.Second,
			},
			want: "'non-default-record' <user> (1s)",
		}, {
			in: ExplanationRecord{
				Name:     "default-record",
				Duration: time.Second,
				Default:  true,
			},
			want: "'default-record' <default> (1s)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.want, tc.in.String())
		})
	}
}

func TestCompileRecord(t *testing.T) {
	tests := []struct {
		description string
		in          Explanation
		add         string
		Default     bool
		want        Explanation
	}{
		{
			description: "basic one step test",
			in:          Explanation{},
			add:         "one",
			want: Explanation{
				Records: []ExplanationRecord{
					{
						Name:     "one",
						Duration: 10 * time.Second,
					},
				},
			},
		},
		{
			description: "basic two step test",
			in: Explanation{
				Records: []ExplanationRecord{
					{
						Name:     "one",
						Duration: 5 * time.Second,
					},
				},
			},
			add: "two",
			want: Explanation{
				Records: []ExplanationRecord{
					{
						Name:     "one",
						Duration: 5 * time.Second,
					},
					{
						Name:     "two",
						Duration: 5 * time.Second,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			start := time.Now()
			end := start.Add(10 * time.Second)

			tc.in.CompileStartedAt = start
			tc.want.CompileStartedAt = start
			tc.in.compileRecord(tc.add, tc.Default, end)

			assert.Equal(tc.want, tc.in)
		})
	}
}
