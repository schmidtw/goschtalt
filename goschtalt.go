// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"sort"
	"sync"

	"github.com/schmidtw/goschtalt/internal/encoding"
	"github.com/schmidtw/goschtalt/internal/encoding/json"
	"github.com/schmidtw/goschtalt/internal/natsort"
)

type raw struct {
	file   string
	values *map[string]any
	sorter Sorter
}

type rawSorter []raw

func (r rawSorter) Len() int           { return len(r) }
func (r rawSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r rawSorter) Less(i, j int) bool { return r[i].sorter(r[i].file, r[j].file) }

// Goschtalt is a configurable, prioritized, merging configuration registry.
type Goschtalt struct {
	codecs *encoding.Registry
	groups []Group
	mutex  sync.Mutex
	sorter Sorter
}

// Option is the type used for options.
type Option func(g *Goschtalt) error

func (fn Option) apply(g *Goschtalt) error {
	return fn(g)
}

// Sorter is the sorting function used to prioritize the configuration files.
type Sorter func(a, b string) bool

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Goschtalt, error) {
	r, _ := encoding.NewRegistry(
		encoding.WithCodec(json.Codec{}))
	g := &Goschtalt{
		codecs: r,
	}

	_ = SortByNatural().apply(g)

	err := g.Options(opts...)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// Option takes a list of options and applies them.
func (g *Goschtalt) Options(opts ...Option) error {
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
func (g *Goschtalt) ReadInConfig() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	full, err := g.readAll()
	if err != nil {
		return err
	}

	return g.merge(full)
}

func (g *Goschtalt) readAll() ([]raw, error) {
	full := []raw{}

	for _, group := range g.groups {
		cfgs, err := group.walk(g.codecs)
		if err != nil {
			return nil, err
		}

		full = append(full, cfgs...)
	}

	// Set the sorter so the list can be properly sorted.
	for i := range full {
		full[i].sorter = g.sorter
	}
	sort.Sort(rawSorter(full))

	return full, nil
}

func (g *Goschtalt) merge(full []raw) error {
	return nil
}

// WithCodec registers a Codec for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func WithCodec(enc encoding.Codec) Option {
	return func(g *Goschtalt) error {
		opt := encoding.WithCodec(enc)
		return g.codecs.Options(opt)
	}
}

// WithoutExtensions provides a mechanism for effectively removing the codecs
// from use for specific file types.
func WithoutExtensions(exts ...string) Option {
	return func(g *Goschtalt) error {
		opt := encoding.WithoutExtensions(exts...)
		return g.codecs.Options(opt)
	}
}

// WithFileGroup provides a group of files, directories or both to examine for
// configuration files.
func WithFileGroup(group Group) Option {
	return func(g *Goschtalt) error {
		g.groups = append(g.groups, group)
		return nil
	}
}

// SortByCustom provides a way to specify your own file sorting logic.  The two
// strings provided are the base filenames.  No directory information is provided.
// For the file 'etc/foo/bar.json' the string given to the sorter will be 'bar.json'.
func SortByCustom(sorter Sorter) Option {
	return func(g *Goschtalt) error {
		g.sorter = sorter
		return nil
	}
}

// SortByLexical provides a simple lexical based sorter for the files where the
// configuration values originate.  This order determines which configuration
// values are adopted first and last.
func SortByLexical() Option {
	return SortByCustom(func(a, b string) bool {
		return a < b
	})
}

// SortByNatural provides a simple lexical based sorter for the files where the
// configuration values originate.  This order determines which configuration
// values are adopted first and last.
func SortByNatural() Option {
	return SortByCustom(natsort.Compare)
}
