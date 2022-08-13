// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"reflect"
)

// typeMapper is a function that maps from one data type to another data type
// if possible, or returns an error if not.
type typeMapper func(any) (any, error)

// CustomMapper provides a way for clients of this library to map from one
// data type to another.  The typ value specifies the destination type the
// mapper provides.  The mappers are called when the Fetch function is called.
// Note it is this function:
//
//	func Fetch[T any](c *Config, key string, want T) (T, error)
func CustomMapper(typ any, fn func(any) (any, error)) Option {
	return func(c *Config) error {
		key := reflect.TypeOf(typ).String()

		if fn == nil {
			delete(c.typeMappers, key)
		} else {
			c.typeMappers[key] = fn
		}
		return nil
	}
}

/*
Here's how to add a duration mapper based on spf13/cast:

import "github.com/spf13/cast"

func WithDurationMapper() Option {
	var d time.Duration
	return WithCustomMapper(d, func(i any) (any, error) {
		return cast.ToDurationE(i)
	})
}
*/
