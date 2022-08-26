// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// decoderRegistry contains the mapping of extension to decoders and encoders.
type decoderRegistry struct {
	decoders map[string]decoder.Decoder

	mutex sync.Mutex
}

// newDecoderRegistry creates a new instance of a decoderRegistry.
func newDecoderRegistry() *decoderRegistry {
	return &decoderRegistry{
		decoders: make(map[string]decoder.Decoder),
	}
}

// extensions returns a list of supported extensions.  Note that all extensions
// are lowercase.
func (dr *decoderRegistry) extensions() (list []string) {
	dr.mutex.Lock()
	decoders := dr.decoders
	dr.mutex.Unlock()

	for key := range decoders {
		list = append(list, key)
	}

	sort.Strings(list)

	return list
}

// find the decoder of interest and return it.
func (dr *decoderRegistry) find(ext string) (decoder.Decoder, error) {
	ext = strings.ToLower(ext)

	dr.mutex.Lock()
	dec, ok := dr.decoders[ext]
	dr.mutex.Unlock()

	if !ok {
		return nil, fmt.Errorf("extension '%s' %w", ext, ErrCodecNotFound)
	}

	return dec, nil
}

// decode decodes based on the specified extension.
func (dr *decoderRegistry) decode(ext, file, keyDelimiter string, b []byte, o *meta.Object) error {
	dec, err := dr.find(ext)
	if err != nil {
		return err
	}

	ctx := decoder.Context{
		Filename:  file,
		Delimiter: keyDelimiter,
	}
	err = dec.Decode(ctx, b, o)
	if err != nil {
		return fmt.Errorf("decoder error for extension '%s' processing file '%s' %w %v",
			ext, file, ErrDecoding, err)
	}
	return nil
}

// register registers a decoder.Decoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func (dr *decoderRegistry) register(enc decoder.Decoder) error {
	normalized := make(map[string]bool)

	exts := enc.Extensions()
	for _, ext := range exts {
		ext = strings.ToLower(ext)
		if _, found := normalized[ext]; found {
			return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
		}
		normalized[ext] = true
	}

	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	for ext := range normalized {
		if _, ok := dr.decoders[ext]; ok {
			return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
		}
	}

	for ext := range normalized {
		dr.decoders[ext] = enc
	}

	return nil
}

// deregister provides a mechanism for effectively removing the decoders from use
// for specific file types.
func (dr *decoderRegistry) deregister(exts ...string) {
	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	for _, ext := range exts {
		delete(dr.decoders, strings.ToLower(ext))
	}

}
