/**
 *  Copyright (c) 2022  Weston Schmidt
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package gestalt

import (
	"sync"

	"github.com/schmidtw/gestalt/internal/encoding"
	"github.com/schmidtw/gestalt/internal/encoding/json"
)

type raw struct {
	file   string
	values *map[string]any
}

// Gestalt is a configurable, prioritized, merging configuration registry.
type Gestalt struct {
	codecs *encoding.Registry
	groups []Group
	mutex  sync.Mutex
}

// Option is the type used for options.
type Option func(g *Gestalt) error

func (fn Option) apply(g *Gestalt) error {
	return fn(g)
}

// New creates a new gestalt configuration instance.
func New(opts ...Option) (*Gestalt, error) {
	r, _ := encoding.NewRegistry(encoding.WithCodec(json.Codec{}))
	g := &Gestalt{
		codecs: r,
	}

	err := g.Options(opts...)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// Option takes a list of options and applies them.
func (g *Gestalt) Options(opts ...Option) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for _, opt := range opts {
		if err := opt.apply(g); err != nil {
			return err
		}
	}
	return nil
}

// ReadInConfig reads in all the files configured using the options provided,
// and merges the configuration trees into a single map for later use.
func (g *Gestalt) ReadInConfig() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	full, err := g.readAll()
	if err != nil {
		return err
	}

	return g.merge(full)
}

func (g *Gestalt) readAll() (map[string]raw, error) {
	full := make(map[string]raw)

	for _, group := range g.groups {
		cfgs, err := group.walk(g.codecs)
		if err != nil {
			return nil, err
		}

		for _, cfg := range cfgs {
			name := cfg.file
			full[name] = cfg
		}
	}

	return full, nil
}

func (g *Gestalt) merge(full map[string]raw) error {
	return nil
}

// WithCodec registers a Codec for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func WithCodec(enc encoding.Codec) Option {
	return func(g *Gestalt) error {
		opt := encoding.WithCodec(enc)
		return g.codecs.Options(opt)
	}
}

// WithoutExtensions provides a mechanism for effectively removing the codecs
// from use for specific file types.
func WithoutExtensions(exts ...string) Option {
	return func(g *Gestalt) error {
		opt := encoding.WithoutExtensions(exts...)
		return g.codecs.Options(opt)
	}
}

// WithFileGroup provides a group of files, directories or both to examine for
// configuration files.
func WithFileGroup(group Group) Option {
	return func(g *Gestalt) error {
		g.groups = append(g.groups, group)
		return nil
	}
}
