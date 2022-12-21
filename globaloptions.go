// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/goschtalt/goschtalt/internal/print"
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

func (errorOption) ignoreDefaults() bool {
	return false
}

func (eo errorOption) apply(*options) error                   { return eo.err }
func (eo errorOption) bufferApply(*bufferOptions) error       { return eo.err }
func (eo errorOption) expandApply(*expand) error              { return eo.err }
func (eo errorOption) marshalApply(*marshalOptions) error     { return eo.err }
func (eo errorOption) unmarshalApply(*unmarshalOptions) error { return eo.err }
func (eo errorOption) valueApply(*valueOptions) error         { return eo.err }

func (eo errorOption) String() string {
	return print.P("WithError", print.Error(eo.err))
}
