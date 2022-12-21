// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/goschtalt/goschtalt/internal/print"
)

// BufferValueOption can be used as a BufferOption or a ValueOption.
type BufferValueOption interface {
	fmt.Stringer

	BufferOption
	ValueOption
}

// AsDefault specifies that this value is a default value & is applied prior to
// any other configuration values.  Default values are applied in the order the
// options are specified.
//
// The unused bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
func AsDefault(asDefault ...bool) BufferValueOption {
	asDefault = append(asDefault, true)

	return optionalAsDefault(asDefault[0])
}

type optionalAsDefault bool

func (o optionalAsDefault) bufferApply(opts *bufferOptions) error {
	opts.isDefault = bool(o)
	return nil
}

func (o optionalAsDefault) valueApply(opts *valueOptions) error {
	opts.isDefault = bool(o)
	return nil
}

func (o optionalAsDefault) String() string {
	return print.P("AsDefault", print.BoolSilentTrue(bool(o)), print.SubOpt())
}
