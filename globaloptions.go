// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/mitchellh/mapstructure"
)

// GlobalOption
type GlobalOption interface {
	fmt.Stringer

	BufferOption
	DecoderConfigOption
	ExpandOption
	MarshalOption
	Option
	ValueOption
	UnmarshalOption
}

// WithError provides a way for plugins to return an error during option
// processing.  This option will always produce the specified error; including
// if the err value is nil.
func WithError(err error) GlobalOption {
	return errorOption{err: err}
}

type errorOption struct {
	err error
}

var _ Option = (*errorOption)(nil)

func (opt errorOption) apply(_ *options) error {
	return opt.err
}

func (errorOption) ignoreDefaults() bool {
	return false
}
func (errorOption) isDefault() bool                            { return false }
func (errorOption) decoderApply(_ *mapstructure.DecoderConfig) {}
func (errorOption) valueApply(_ *valueOptions)                 {}
func (errorOption) expandApply(*expand)                        {}
func (errorOption) marshalApply(*marshalOptions)               {}
func (errorOption) unmarshalApply(*unmarshalOptions)           {}

func (o errorOption) String() string {
	return print.P("WithError", print.Error(o.err))
}
