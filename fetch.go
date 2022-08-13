// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/schmidtw/goschtalt/pkg/meta"
)

// Fetch provides a generic based strict typed approach to fetching parts of the
// configuration tree.  The Config and key parameters are fairly
// straighforward, but the want may not be.  The want parameter is used to
// determine the type of the output object desired.  This allows this function
// to do to handy things the c.Fetch() method can't do:
//
//  - Thes function is able to validate the type returned is the type desired,
//    or return a descriptive error about why it can't do what was asked for.
//
//  - This function is also able to perform remapping from an existing type to
//    what you want based on the typeMappers provided.  This allows you to
//    automatically convert and cast a string to a time.Duration if you provide
//    the mapper.
//
// Here is an example showing how to add a duration caster based on spf13/cast:
//
//    import (
//        "github.com/schmidtw/goschtalt"
//        "github.com/spf13/cast"
//    )
//
//    func DurationMapper() goschtalt.Option {
//        var d time.Duration
//        return goschtalt.CustomMapper(d, func(i any) (any, error) {
//            return cast.ToDurationE(i)
//        })
//    }
//
//    ...
//
//    c := goschtalt.New(DurationMapper())
//
//    ...
func Fetch[T any](c *Config, key string, want T) (T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var zeroVal T
	rv, _, err := c.fetchWithOrigin(key)
	if err != nil {
		return zeroVal, err
	}

	if fn, found := c.typeMappers[reflect.TypeOf(want).String()]; found {
		rv, err = fn(rv)
		if err != nil {
			return zeroVal, err
		}
	}

	if reflect.TypeOf(want) != reflect.TypeOf(rv) {
		return zeroVal, fmt.Errorf("%w: expected type '%s' does not match type found '%s'",
			ErrTypeMismatch,
			reflect.TypeOf(want),
			reflect.TypeOf(rv))
	}

	return rv.(T), nil
}

// Fetch pulls the specified portion of the configuration tree and returns it to
// the caller as an any, since it could be a map node or a specific value.
func (c *Config) Fetch(key string) (any, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	a, _, err := c.fetchWithOrigin(key)
	return a, err
}

// FetchWithOrigin pulls the specified portion of the configuration tree as
// well as the origin information and returns both sets of data to the caller
// as an any and array of meta.Origin values.
func (c *Config) FetchWithOrigin(key string) (any, []meta.Origin, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.fetchWithOrigin(key)
}

// fetchWithOrigin is the non-mutex wrapped function that does the work, but
// can be called by the other mutex protected functions without deadlocking.
func (c *Config) fetchWithOrigin(key string) (any, []meta.Origin, error) {
	if !c.hasBeenCompiled {
		return nil, nil, ErrNotCompiled
	}

	if len(key) == 0 {
		return c.tree.ToRaw(), c.tree.Origins, nil
	}

	key = c.keySwizzler(key)
	path := strings.Split(key, c.keyDelimiter)

	obj, err := c.tree.Fetch(path, c.keyDelimiter)
	if err != nil {
		return nil, nil, err
	}

	return obj.ToRaw(), obj.Origins, nil
}
