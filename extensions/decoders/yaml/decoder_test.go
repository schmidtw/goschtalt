// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

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

	want := []string{"yaml", "yml"}
	got := Decoder{}.Extensions()

	assert.Empty(cmp.Diff(want, got))
}

func TestDecode(t *testing.T) {
	unknownErr := errors.New("unknown error")

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
			description: "A small test.",
			in: `---
a:
  b:
    c: '123'
d:
  e:
    - fog
    - dog`,
			expected: meta.Object{
				Origins: []meta.Origin{{File: "file.yml", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"a": {
						Origins: []meta.Origin{{File: "file.yml", Line: 2, Col: 1}},
						Map: map[string]meta.Object{
							"b": {
								Origins: []meta.Origin{{File: "file.yml", Line: 3, Col: 3}},
								Map: map[string]meta.Object{
									"c": {
										Origins: []meta.Origin{{File: "file.yml", Line: 4, Col: 8}},
										Value:   "123",
									},
								},
							},
						},
					},
					"d": {
						Origins: []meta.Origin{{File: "file.yml", Line: 5, Col: 1}},
						Map: map[string]meta.Object{
							"e": {
								Origins: []meta.Origin{{File: "file.yml", Line: 6, Col: 3}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 7, Col: 7}},
										Value:   "fog",
									},
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 8, Col: 7}},
										Value:   "dog",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "An anchor and merge ... which fails gracefully.",
			in: `---
a:
  b: &foo
    c: '123'
  z: &rat
    y: cat
d:
  <<: *foo
  e: &bar
    - fog
    - dog
  f: &car
    - red
    - blue
g:
  <<: [ *foo, *rat ]
h:
  [*bar, *car]
`,
			expected: meta.Object{
				Origins: []meta.Origin{{File: "file.yml", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"a": {
						Origins: []meta.Origin{{File: "file.yml", Line: 2, Col: 1}},
						Map: map[string]meta.Object{
							"b": {
								Origins: []meta.Origin{{File: "file.yml", Line: 3, Col: 3}},
								Map: map[string]meta.Object{
									"c": {
										Origins: []meta.Origin{{File: "file.yml", Line: 4, Col: 8}},
										Value:   "123",
									},
								},
							},
							"z": {
								Origins: []meta.Origin{{File: "file.yml", Line: 5, Col: 3}},
								Map: map[string]meta.Object{
									"y": {
										Origins: []meta.Origin{{File: "file.yml", Line: 6, Col: 8}},
										Value:   "cat",
									},
								},
							},
						},
					},
					"d": {
						Origins: []meta.Origin{{File: "file.yml", Line: 7, Col: 1}},
						Map: map[string]meta.Object{
							"c": {
								Origins: []meta.Origin{},
								Value:   "123",
							},
							"e": {
								Origins: []meta.Origin{{File: "file.yml", Line: 9, Col: 3}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 10, Col: 7}},
										Value:   "fog",
									},
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 11, Col: 7}},
										Value:   "dog",
									},
								},
							},
							"f": {
								Origins: []meta.Origin{{File: "file.yml", Line: 12, Col: 3}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 13, Col: 7}},
										Value:   "red",
									},
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 14, Col: 7}},
										Value:   "blue",
									},
								},
							},
						},
					},
					"g": {
						Origins: []meta.Origin{{File: "file.yml", Line: 15, Col: 1}},
						Map: map[string]meta.Object{
							"c": {
								Origins: []meta.Origin{},
								Value:   "123",
							},
							"y": {
								Origins: []meta.Origin{},
								Value:   "cat",
							},
						},
					},
					"h": {
						Origins: []meta.Origin{{File: "file.yml", Line: 17, Col: 1}},
						Array: []meta.Object{
							{
								Origins: []meta.Origin{{File: "file.yml", Line: 18, Col: 4}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 10, Col: 7}},
										Value:   "fog",
									},
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 11, Col: 7}},
										Value:   "dog",
									},
								},
							},
							{
								Origins: []meta.Origin{{File: "file.yml", Line: 18, Col: 10}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 13, Col: 7}},
										Value:   "red",
									},
									{
										Origins: []meta.Origin{{File: "file.yml", Line: 14, Col: 7}},
										Value:   "blue",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "An invalid yaml file",
			in:          `this is invalid=yaml`,
			expectedErr: unknownErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var d Decoder
			var got meta.Object
			ctx := decoder.Context{
				Filename:  "file.yml",
				Delimiter: ".",
			}
			err := d.Decode(ctx, []byte(tc.in), &got)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
				return
			}

			if errors.Is(unknownErr, tc.expectedErr) {
				assert.Error(err)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
