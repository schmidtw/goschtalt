// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

func TestExtensions(t *testing.T) {
	assert := assert.New(t)

	want := []string{"json"}
	got := Decoder{}.Extensions()

	assert.Empty(cmp.Diff(want, got))
}

func TestDecode(t *testing.T) {
	unknown := errors.New("unknown")
	tests := []struct {
		description string
		in          string
		expected    meta.Object
		expectedErr error
	}{
		{
			description: "A test of empty.",
			expected:    meta.Object{},
		}, {
			description: "Invalid json.",
			in:          `{ a b }`,
			expectedErr: unknown,
		}, {
			description: "A small test.",
			in:          `{ "a": { "b": { "c": "123" } }, "d": { "e": [ "fog", "dog" ] } }`,
			expected: meta.Object{
				Origins: []meta.Origin{{File: "file.json"}},
				Map: map[string]meta.Object{
					"a": {
						Origins: []meta.Origin{{File: "file.json"}},
						Map: map[string]meta.Object{
							"b": {
								Origins: []meta.Origin{{File: "file.json"}},
								Map: map[string]meta.Object{
									"c": {
										Origins: []meta.Origin{{File: "file.json"}},
										Value:   "123",
									},
								},
							},
						},
					},
					"d": {
						Origins: []meta.Origin{{File: "file.json"}},
						Map: map[string]meta.Object{
							"e": {
								Origins: []meta.Origin{{File: "file.json"}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.json"}},
										Value:   "fog",
									},
									{
										Origins: []meta.Origin{{File: "file.json"}},
										Value:   "dog",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var d Decoder
			var got meta.Object
			ctx := decoder.Context{
				Filename:  "file.json",
				Delimiter: ".",
			}
			err := d.Decode(ctx, []byte(tc.in), &got)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
			}

			if tc.expectedErr == unknown {
				assert.NotNil(err)
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
