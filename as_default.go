// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// AsDefaultOption
type AsDefaultOption interface {
	fmt.Stringer

	ValueOption
	BufferOption
}

// AsDefault specifies that this value is a default value & is applied prior to
// any other configuration values.  Default values are applied in the order the
// options are specified.
func AsDefault(asDefault ...bool) AsDefaultOption {
	asDefault = append(asDefault, true)

	return optionalAsDefault(asDefault[0])
}

type optionalAsDefault bool

func (o optionalAsDefault) isDefault() bool {
	return bool(o)
}

func (_ optionalAsDefault) decoderApply(_ *mapstructure.DecoderConfig) {}
func (_ optionalAsDefault) valueApply(_ *valueOptions)                 {}

func (o optionalAsDefault) String() string {
	if o {
		return "AsDefault()"
	}
	return "AsDefault(false)"
}
