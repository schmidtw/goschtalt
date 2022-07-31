// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"reflect"
	"sync"

	"github.com/schmidtw/goschtalt/internal/encoding"
	"github.com/schmidtw/goschtalt/internal/encoding/json"
)

var (
	ErrConflict      = errors.New("a conflict has been detected")
	ErrInvalidOption = errors.New("invalid option")
)

type raw struct {
	file   string
	config *map[string]any
}

type annotatedMap struct {
	files []string
	m     map[string]any
}

type annotatedArray struct {
	files []string
	array []any
}

type annotatedValue struct {
	files []string
	value any
}

// Goschtalt is a configurable, prioritized, merging configuration registry.
type Goschtalt struct {
	codecs          *encoding.Registry
	groups          []Group
	annotated       annotatedMap
	mutex           sync.Mutex
	rawSorter       func([]raw)
	keySwizzler     func(*map[string]any)
	leafConflictFn  func(cur, next annotatedValue) (annotatedValue, error)
	arrayConflictFn func(cur, next annotatedArray) (annotatedArray, error)
	mapConflictFn   func(any, any) (any, error)
}

// Option is the type used for options.
type Option func(g *Goschtalt) error

func (fn Option) apply(g *Goschtalt) error {
	return fn(g)
}

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Goschtalt, error) {
	r, _ := encoding.NewRegistry()
	g := &Goschtalt{
		codecs: r,
	}

	/* set the defaults */
	_ = g.Options(
		WithCodec(json.Codec{}),
		WithSortOrder(Natural),
		WithKeyCase(Lower),
		WithMergeStrategy(Map, Fail),
		WithMergeStrategy(Array, Append),
		WithMergeStrategy(Value, Latest),
	)

	/* apply the specified options */
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

func (g *Goschtalt) readAll() ([]annotatedMap, error) {
	full := []raw{}

	for _, group := range g.groups {
		cfgs, err := group.walk(g.codecs)
		if err != nil {
			return nil, err
		}

		full = append(full, cfgs...)
	}

	for i := range full {
		// Apply any key mangling that is needed.
		g.keySwizzler(full[i].config)
	}
	g.rawSorter(full)

	nodes := []annotatedMap{}

	for _, cfg := range full {
		n := rawToAnnotatedMap(cfg.file, *cfg.config)
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func rawToAnnotatedMap(file string, src map[string]any) annotatedMap {
	m := annotatedMap{
		files: []string{file},
	}

	tmp := map[string]any{}
	for key, val := range src {
		tmp[key] = rawToAnnotatedVal(file, val)
	}
	m.m = tmp

	return m
}

func rawToAnnotatedVal(file string, val any) any {
	switch val := val.(type) {
	case map[string]any:
		return rawToAnnotatedMap(file, val)
	case []any:
		return rawToAnnotatedArray(file, val)
	}

	return annotatedValue{
		files: []string{file},
		value: val,
	}
}

func rawToAnnotatedArray(file string, a []any) annotatedArray {
	rv := annotatedArray{
		files: []string{file},
		array: make([]any, len(a)),
	}
	for i, val := range a {
		rv.array[i] = rawToAnnotatedVal(file, val)
	}

	return rv
}

func (g *Goschtalt) merge(cfgs []annotatedMap) error {
	if len(cfgs) == 0 {
		return nil
	}

	g.annotated = cfgs[0]
	for _, cfg := range cfgs[1:] {
		if err := g.mergeMap(cfg, &g.annotated); err != nil {
			return err
		}
	}
	return nil
}

func (g *Goschtalt) mergeMap(src annotatedMap, dest *annotatedMap) error {
	for key, next := range src.m {
		current, found := dest.m[key]
		if !found {
			// No conflicts - easy merge
			switch next := next.(type) {
			case annotatedMap:
				dest.files = dedupedAppend(dest.files, next.files...)
			case annotatedArray:
				dest.files = dedupedAppend(dest.files, next.files...)
			case annotatedValue:
				dest.files = dedupedAppend(dest.files, next.files...)
			}
			dest.m[key] = next
			continue
		}

		if reflect.TypeOf(current) == reflect.TypeOf(next) {
			// Types match.
			switch next := next.(type) {
			case annotatedMap:
				// This one is a bit of a special case in that there isn't a
				// conflict as much as we ran into another structural map.
				// Because of that, the conflict resolver doesn't get called
				// here.
				dest.files = dedupedAppend(dest.files, next.files...)
				c := current.(annotatedMap)
				if err := g.mergeMap(next, &c); err != nil {
					return err
				}
				dest.m[key] = c
			case annotatedArray: // Both arrays, resolve.
				dest.files = dedupedAppend(dest.files, next.files...)
				tmp, err := g.arrayConflictFn(current.(annotatedArray), next)
				if err != nil {
					return err
				}
				dest.m[key] = tmp
			case annotatedValue: // Both values, resolve.
				dest.files = dedupedAppend(dest.files, next.files...)
				tmp, err := g.leafConflictFn(current.(annotatedValue), next)
				if err != nil {
					return err
				}
				dest.m[key] = tmp
			}
			continue
		}

		// The types don't match, resolve.
		tmp, err := g.mapConflictFn(current, next)
		if err != nil {
			return err
		}

		// Adjust the file metrics
		switch tmp := tmp.(type) {
		case annotatedMap:
			dest.files = dedupedAppend(dest.files, tmp.files...)
		case annotatedArray:
			dest.files = dedupedAppend(dest.files, tmp.files...)
		case annotatedValue:
			dest.files = dedupedAppend(dest.files, tmp.files...)
		}
		dest.m[key] = tmp
	}

	return nil
}

func dedupedAppend(list []string, added ...string) []string {
	keys := make(map[string]bool)
	for _, item := range list {
		keys[item] = true
	}

	for _, want := range added {
		if _, found := keys[want]; !found {
			keys[want] = true
			list = append(list, want)
		}
	}
	return list
}
