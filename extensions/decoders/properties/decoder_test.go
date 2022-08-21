// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package properties

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

func TestExtensions(t *testing.T) {
	assert := assert.New(t)

	want := []string{"properties"}
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
			description: "A test of invalid values.",
			in:          `=:`,
			expectedErr: unknown,
		}, {
			description: "A small test.",
			in: `# A small test
Foo.bar.0 = hello
Foo.bar.1 = cat
Foo.bar.2 = bat
Foo.ba = sheep
Foo.d = milk
Foo.l = jestor`,
			expected: meta.Object{
				Origins: []meta.Origin{{File: "file.properties", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"Foo": {
						Origins: []meta.Origin{{File: "file.properties", Line: 2, Col: 1}},
						Map: map[string]meta.Object{
							"ba": {
								Origins: []meta.Origin{{File: "file.properties", Line: 5, Col: 1}},
								Value:   "sheep",
							},
							"bar": {
								Origins: []meta.Origin{{File: "file.properties", Line: 2, Col: 1}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.properties", Line: 2, Col: 1}},
										Value:   "hello",
									},
									{
										Origins: []meta.Origin{{File: "file.properties", Line: 3, Col: 1}},
										Value:   "cat",
									},
									{
										Origins: []meta.Origin{{File: "file.properties", Line: 4, Col: 1}},
										Value:   "bat",
									},
								},
							},
							"d": {
								Origins: []meta.Origin{{File: "file.properties", Line: 6, Col: 1}},
								Value:   "milk",
							},
							"l": {
								Origins: []meta.Origin{{File: "file.properties", Line: 7, Col: 1}},
								Value:   "jestor",
							},
						},
					},
				},
			},
		}, {
			description: "A test of types.",
			in: `# A test of types.
a = 250
b = 0xff
c = 077
d = 13.2
e = true
f = false`,
			expected: meta.Object{
				Origins: []meta.Origin{{File: "file.properties", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"a": {
						Origins: []meta.Origin{{File: "file.properties", Line: 2, Col: 1}},
						Value:   int64(250),
					},
					"b": {
						Origins: []meta.Origin{{File: "file.properties", Line: 3, Col: 1}},
						Value:   int64(255),
					},
					"c": {
						Origins: []meta.Origin{{File: "file.properties", Line: 4, Col: 1}},
						Value:   int64(63),
					},
					"d": {
						Origins: []meta.Origin{{File: "file.properties", Line: 5, Col: 1}},
						Value:   float64(13.2),
					},
					"e": {
						Origins: []meta.Origin{{File: "file.properties", Line: 6, Col: 1}},
						Value:   true,
					},
					"f": {
						Origins: []meta.Origin{{File: "file.properties", Line: 7, Col: 1}},
						Value:   false,
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
			err := d.Decode("file.properties", ".", []byte(tc.in), &got)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
				return
			}

			if tc.expectedErr == unknown {
				assert.Error(err)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
