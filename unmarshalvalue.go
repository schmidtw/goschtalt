// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/goschtalt/goschtalt/internal/print"
)

// UnmarshalValueOption options are options shared between UnmarshalOption and
// ValueOption interfaces.
type UnmarshalValueOption interface {
	fmt.Stringer

	UnmarshalOption
	ValueOption
}

// TagName defines which tag goschtalt honors when it unmarshals to/from
// structures.  The name string defines the new tag name to read.  If an empty
// string is passed in then the module default will be used.
//
// # Default
//
// "goschtalt"
func TagName(name string) UnmarshalValueOption {
	return tagNameOption(name)
}

type tagNameOption string

func (val tagNameOption) unmarshalApply(opts *unmarshalOptions) error {
	tag := defaultTag
	if len(string(val)) > 0 {
		tag = string(val)
	}
	opts.decoder.TagName = tag
	return nil
}

func (val tagNameOption) valueApply(opts *valueOptions) error {
	tag := defaultTag
	if len(string(val)) > 0 {
		tag = string(val)
	}
	opts.tagName = tag
	return nil
}

func (val tagNameOption) String() string {
	return print.P("TagName", print.String(string(val)), print.SubOpt())
}

// Mapper takes a golang structure field string and outputs a goschtalt
// configuration tree name string that is one of the following:
//   - "" indicating this mapper was unable to perform the remapping, continue
//     calling mappers in the chain
//   - "-"  indicating this value should be dropped entirely
//   - anything else indicates the new full name
type Mapper func(s string) string

// Keymap takes a map of strings to strings and adds it to the existing
// chain of keymaps. The key of the map is the golang structure field name and
// the value is the goschtalt configuration tree name string. The value of "-"
// means do not convert, and an empty string means call the next in the chain.
//
// For example, the map below converts a structure field "FooBarIP" to "foobar_ip".
//
//	Keymap( map[string]string{
//		"FooBarIP": "foobar_ip",
//	})
func Keymap(m map[string]string) UnmarshalValueOption {
	return &keymapOption{
		text: print.P("Keymap", print.StringMap(m), print.SubOpt()),
		m: func(s string) string {
			if val, found := m[s]; found {
				return val
			}
			return s
		},
	}
}

// KeymapFn takes a Mapper function and adds it to the existing chain of
// mappers, in the front of the list.
//
// This allows for multiple mappers to be specified instead of requiring a
// single mapper with full knowledge of how to map everything. This makes it
// easy to add logic to remap full keys without needing to re-implement the
// underlying converters.
func KeymapFn(mapper Mapper) UnmarshalValueOption {
	return &keymapOption{
		text: print.P("KeymapFn", print.Fn(mapper), print.SubOpt()),
		m:    mapper,
	}
}

type keymapOption struct {
	text string
	m    Mapper
}

func (k keymapOption) unmarshalApply(opts *unmarshalOptions) error {
	if k.m != nil {
		opts.mappers = append([]Mapper{k.m}, opts.mappers...)
	}
	return nil
}

func (k keymapOption) valueApply(opts *valueOptions) error {
	if k.m != nil {
		opts.mappers = append([]Mapper{k.m}, opts.mappers...)
	}
	return nil
}

func (k keymapOption) String() string {
	return k.text
}
