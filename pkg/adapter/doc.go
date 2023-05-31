// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// Provides adapters for built-in types and interfaces.
//
// The majority of the adapters are simple string to type style converters.
// [BoolUnmarshal] and [MarshalBool] are examples of the simple converters.
//
// There is also a special adapter pair that enable the [encoding.TextMarshaler]
// and [encoding.TextUnmarshaler] interfaces:  [TextUnmarshal]() and [MarshalText]().
//
// Both [TextUnmarshal] and [MarshalText] take a [Matcher] that allows you to
// control which objects should use their provided functions.  Generally you
// will want to use the provided methods, but sometimes they either don't work
// or don't fit the use case right and replacing them is desired.  A Matcher is
// how you can block the provided methods.
package adapter
