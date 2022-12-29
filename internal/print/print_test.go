// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package print

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP(t *testing.T) {
	tests := []struct {
		noOpts bool
		opt    Option
		opts   []Option
		expect string
	}{
		{
			expect: "Foo()",
		}, {
			expect: "Foo()",
			noOpts: true,
		}, {
			opt:    Bool(false),
			expect: "Foo( false )",
		}, {
			opt:    Bool(false, "label"),
			expect: "Foo( label: false )",
		}, {
			opt:    Bool(true),
			expect: "Foo( true )",
		}, {
			opt:    Bool(true, "label"),
			expect: "Foo( label: true )",
		}, {
			opt:    BoolSilentFalse(false),
			expect: "Foo()",
		}, {
			opt:    BoolSilentFalse(false, "label"),
			expect: "Foo()",
		}, {
			opt:    BoolSilentFalse(true),
			expect: "Foo( true )",
		}, {
			opt:    BoolSilentFalse(true, "label"),
			expect: "Foo( label: true )",
		}, {
			opt:    BoolSilentTrue(false),
			expect: "Foo( false )",
		}, {
			opt:    BoolSilentTrue(false, "label"),
			expect: "Foo( label: false )",
		}, {
			opt:    BoolSilentTrue(true),
			expect: "Foo()",
		}, {
			opt:    BoolSilentTrue(true, "label"),
			expect: "Foo()",
		}, {
			opt:    Bytes(nil),
			expect: "Foo( nil )",
		}, {
			opt:    Bytes(nil, "label"),
			expect: "Foo( label: nil )",
		}, {
			opt:    Bytes([]byte("hi")),
			expect: "Foo( []byte )",
		}, {
			opt:    Bytes([]byte("hi"), "label"),
			expect: "Foo( label: []byte )",
		}, {
			opt:    Error(nil),
			expect: "Foo( nil )",
		}, {
			opt:    Error(nil, "label"),
			expect: "Foo( label: nil )",
		}, {
			opt:    Error(errors.New("test")),
			expect: "Foo( 'test' )",
		}, {
			opt:    Error(errors.New("test"), "label"),
			expect: "Foo( label: 'test' )",
		}, {
			opt:    Fn(nil),
			expect: "Foo( nil )",
		}, {
			opt:    Fn(nil, "label"),
			expect: "Foo( label: nil )",
		}, {
			opt:    Fn(func() {}),
			expect: "Foo( custom )",
		}, {
			opt:    Fn(func() {}, "label"),
			expect: "Foo( label: custom )",
		}, {
			opt:    FnAltNil(nil, "empty"),
			expect: "Foo( empty )",
		}, {
			opt:    FnAltNil(nil, "empty", "label"),
			expect: "Foo( label: empty )",
		}, {
			opt:    FnAltNil(func() {}, "empty"),
			expect: "Foo( custom )",
		}, {
			opt:    FnAltNil(func() {}, "empty", "label"),
			expect: "Foo( label: custom )",
		}, {
			opt:    FnAltNotNil(nil, "something"),
			expect: "Foo( nil )",
		}, {
			opt:    FnAltNotNil(nil, "something", "label"),
			expect: "Foo( label: nil )",
		}, {
			opt:    FnAltNotNil(func() {}, "something"),
			expect: "Foo( something )",
		}, {
			opt:    FnAltNotNil(func() {}, "something", "label"),
			expect: "Foo( label: something )",
		}, {
			opt:    Int(5),
			expect: "Foo( 5 )",
		}, {
			opt:    Int(5, "label"),
			expect: "Foo( label: 5 )",
		}, {
			opt:    Literal("foo"),
			expect: "Foo( foo )",
		}, {
			opt:    Literal("foo", "label"),
			expect: "Foo( label: foo )",
		}, {
			opt:    LiteralStrings([]string{"foo", "bar"}),
			expect: "Foo( foo, bar )",
		}, {
			opt:    LiteralStrings([]string{"foo", "bar"}, "label"),
			expect: "Foo( label: foo, bar )",
		}, {
			opt:    LiteralStringers([]fmt.Stringer{Literal("one"), Literal("two")}),
			expect: "Foo( one, two )",
		}, {
			opt:    LiteralStringers([]fmt.Stringer{Literal("one"), Literal("two")}, "label"),
			expect: "Foo( label: one, two )",
		}, {
			opt:    String("foo"),
			expect: "Foo( 'foo' )",
		}, {
			opt:    String("foo", "label"),
			expect: "Foo( label: 'foo' )",
		}, {
			opt:    Stringers([]fmt.Stringer{Literal("one"), Literal("two"), Literal("three")}),
			expect: "Foo( 'one', 'two', 'three' )",
		}, {
			opt:    Stringers([]fmt.Stringer{Literal("one"), Literal("two"), Literal("three")}, "label"),
			expect: "Foo( label: 'one', 'two', 'three' )",
		}, {
			opt:    Stringers([]fmt.Stringer{}),
			expect: "Foo( '' )",
		}, {
			opt:    Stringers([]fmt.Stringer{}, "label"),
			expect: "Foo( label: '' )",
		}, {
			opt:    Strings([]string{"one", "two", "three"}),
			expect: "Foo( 'one', 'two', 'three' )",
		}, {
			opt:    Strings([]string{"one", "two", "three"}, "label"),
			expect: "Foo( label: 'one', 'two', 'three' )",
		}, {
			opt:    Strings([]string{}),
			expect: "Foo( '' )",
		}, {
			opt:    Strings([]string{}, "label"),
			expect: "Foo( label: '' )",
		}, {
			opt:    StringMap(map[string]string{}),
			expect: "Foo( map[0] )",
		}, {
			opt:    StringMap(map[string]string{}, "label"),
			expect: "Foo( label: map[0] )",
		}, {
			opt: StringMap(map[string]string{
				"dog": "cat",
			}),
			expect: "Foo( map[1] )",
		}, {
			opt: StringMap(map[string]string{
				"dog": "cat",
			}, "label"),
			expect: "Foo( label: map[1] )",
		}, {
			opts: []Option{
				Bool(true), Fn(nil, "label"), Literal("abs"),
			},
			expect: "Foo( true, label: nil, abs )",
		}, {
			opts: []Option{
				Bool(true), nil, Fn(nil, "label"), Literal("abs"),
				Yields(
					Bool(true, "a"), nil, Fn(nil, "b"), Literal("abs", "c"),
				),
			},
			expect: "Foo( true, label: nil, abs ) --> a: true, b: nil, c: abs",
		}, {
			opts: []Option{
				Bool(true, "label"), Fn(nil), Literal("abs"), SubOpt(),
			},
			expect: "Foo(label: true, nil, abs)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.expect, func(t *testing.T) {
			assert := assert.New(t)

			var got string
			if len(tc.opts) > 0 {
				got = P("Foo", tc.opts...)
			} else {
				if tc.noOpts {
					got = P("Foo")
				} else {
					got = P("Foo", tc.opt)
				}
			}

			assert.Equal(tc.expect, got)
		})
	}
}
