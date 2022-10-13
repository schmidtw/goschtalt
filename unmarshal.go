// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// AddDefaultUnmarshalOptions allows customization of the desired options for all
// invocations of the Unmarshal() function.  This should make consistent use
// use of the Unmarshal() call easier.
func AddDefaultUnmarshalOptions(opts ...MapstructureOption) Option {
	return func(c *Config) error {
		c.unmarshalOptions = append(c.unmarshalOptions, opts...)
		return nil
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
func (c *Config) Unmarshal(key string, result any, opts ...MapstructureOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var d decoderConfig
	d.cfg.Result = result

	full := append(c.unmarshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			opt(&d)
		}
	}

	tree, _, err := c.fetchWithOrigin(key)
	if err != nil {
		if errors.Is(err, meta.ErrNotFound) && d.optional {
			return nil
		}
		return err
	}

	decoder, err := mapstructure.NewDecoder(&d.cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(tree)
}

// UnmarshalFn returns a function that takes a goschtalt Config structure and
// returns a function that allows for unmarshalling of a portion of the tree
// specified by the key into a zero value type.
//
// This function is specifically helpful with DI frameworks like Uber's fx
// framework.
//
// In this short example, the type myStruct is created and populated with the
// configuring values found under the "conf" key in the goschtalt configuration.
//
//	app := fx.New(
//		fx.Provide(
//			goschtalt.UnmarshalFn[myStruct]("conf"),
//		),
//	)
func UnmarshalFn[T any](key string, opts ...MapstructureOption) func(*Config) (T, error) {
	return func(cfg *Config) (T, error) {
		return Fetch[T](cfg, key, opts...)
	}
}

// Fetch provides a generics based strict typed approach to fetching parts of the
// configuration tree.
func Fetch[T any](c *Config, key string, opts ...MapstructureOption) (T, error) {
	var rv T
	err := c.Unmarshal(key, &rv, opts...)
	if err != nil {
		var zeroVal T
		return zeroVal, err
	}

	return rv, nil
}
