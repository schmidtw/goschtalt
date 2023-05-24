// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errUnknown = errors.New("errUnknown")
)

type adapterToCfg interface {
	To(from reflect.Value) (any, error)
}

type adapterFromCfg interface {
	From(from, to reflect.Value) (any, error)
}

type valueAdapterTest struct {
	description string
	from        any
	obj         adapterToCfg
	expect      any
	expectErr   error
}

type unmarshalAdapterTest struct {
	description string
	from        any
	to          any
	obj         adapterFromCfg
	expect      any
	expectErr   error
}

func testValueAdapters(t *testing.T, tests []valueAdapterTest) {
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := tc.obj.To(reflect.ValueOf(tc.from))

			if errors.Is(errUnknown, tc.expectErr) {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.Error(err)
				return
			}

			if tc.expectErr != nil {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.ErrorIs(err, tc.expectErr)
				return
			}

			assert.Equal(tc.expect, got)
			assert.NoError(err)
		})
	}
}

func testUnmarshalAdapters(t *testing.T, tests []unmarshalAdapterTest) {
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := tc.obj.From(reflect.ValueOf(tc.from), reflect.ValueOf(tc.to))

			if errors.Is(errUnknown, tc.expectErr) {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.Error(err)
				return
			}

			if tc.expectErr != nil {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.ErrorIs(err, tc.expectErr)
				return
			}

			assert.EqualValues(tc.expect, got)
			assert.NoError(err)
		})
	}
}
