// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package encoding

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	ErrDuplicateFound = errors.New("duplicate found")
	ErrNotFound       = errors.New("not found")
	ErrDecoding       = errors.New("decoding error")
)

// Codec provides the encoder and decoders interface
type Codec interface {
	Encode(v *map[string]any) ([]byte, error)
	Decode(b []byte, v *map[string]any) error
	Extensions() []string
}

// Option is the type used for options.
type Option func(r *Registry) error

func (fn Option) apply(r *Registry) error {
	return fn(r)
}

// Registry contains the mapping of extension to decoders and encoders.
type Registry struct {
	codecs map[string]Codec

	mutex sync.RWMutex
}

// NewRegistry creates a new instance of a Registry.
func NewRegistry(opts ...Option) (*Registry, error) {
	r := &Registry{
		codecs: make(map[string]Codec),
	}

	err := r.Options(opts...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Option takes a list of options and applies them.
func (r *Registry) Options(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.apply(r); err != nil {
			return err
		}
	}
	return nil
}

// Extensions returns a list of supported extensions.  Note that all extensions
// are lowercase.
func (r *Registry) Extensions() (list []string) {
	r.mutex.RLock()
	codecs := r.codecs
	r.mutex.RUnlock()

	for key := range codecs {
		list = append(list, key)
	}

	sort.Strings(list)

	return list
}

// find the codec of interest and return it.
func (r *Registry) find(ext string) (Codec, error) {
	ext = strings.ToLower(ext)

	r.mutex.RLock()
	codec, ok := r.codecs[ext]
	r.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("extension '%s' %w", ext, ErrNotFound)
	}

	return codec, nil
}

// Decode decodes based on the specified extension.
func (r *Registry) Decode(ext string, b []byte, v *map[string]any) error {
	codec, err := r.find(ext)
	if err != nil {
		return err
	}

	err = codec.Decode(b, v)
	if err != nil {
		return fmt.Errorf("%w %v", ErrDecoding, err)
	}
	return nil
}

// Encode encodes based on the specified extension.
func (r *Registry) Encode(ext string, v *map[string]any) ([]byte, error) {
	codec, err := r.find(ext)
	if err != nil {
		return nil, err
	}

	return codec.Encode(v)
}

// WithCodec registers a Codec for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func WithCodec(enc Codec) Option {
	return func(r *Registry) error {
		normalized := make(map[string]bool)

		exts := enc.Extensions()
		for _, ext := range exts {
			ext = strings.ToLower(ext)
			if _, found := normalized[ext]; found {
				return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
			}
			normalized[ext] = true
		}

		r.mutex.Lock()
		defer r.mutex.Unlock()

		for ext := range normalized {
			if _, ok := r.codecs[ext]; ok {
				return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
			}
		}

		for ext := range normalized {
			r.codecs[ext] = enc
		}

		return nil
	}
}

// WithoutExtensions provides a mechanism for effectively removing the codecs
// from use for specific file types.
func WithoutExtensions(exts ...string) Option {
	return func(r *Registry) error {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		for _, ext := range exts {
			delete(r.codecs, strings.ToLower(ext))
		}

		return nil
	}
}
