// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/pkg/encoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// encoderRegistry contains the mapping of extension to encoders and encoders.
type encoderRegistry struct {
	encoders map[string]encoder.Encoder

	mutex sync.RWMutex
}

// newEncoderRegistry creates a new instance of a encoderRegistry.
func newEncoderRegistry() *encoderRegistry {
	return &encoderRegistry{
		encoders: make(map[string]encoder.Encoder),
	}
}

// Extensions returns a list of supported extensions.  Note that all extensions
// are lowercase.
func (er *encoderRegistry) extensions() (list []string) {
	er.mutex.RLock()
	encoders := er.encoders
	er.mutex.RUnlock()

	for key := range encoders {
		list = append(list, key)
	}

	sort.Strings(list)

	return list
}

// find the codec of interest and return it.
func (er *encoderRegistry) find(ext string) (encoder.Encoder, error) {
	ext = strings.ToLower(ext)

	er.mutex.RLock()
	encoder, ok := er.encoders[ext]
	er.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("extension '%s' %w", ext, ErrCodecNotFound)
	}

	return encoder, nil
}

// encode based on the specified extension.
func (er *encoderRegistry) encode(ext string, v any) ([]byte, error) {
	encoder, err := er.find(ext)
	if err != nil {
		return nil, err
	}

	b, err := encoder.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("%w %v", ErrEncoding, err)
	}
	return b, nil
}

// encodeExtended based on the specified extension.
func (er *encoderRegistry) encodeExtended(ext string, o meta.Object) ([]byte, error) {
	encoder, err := er.find(ext)
	if err != nil {
		return nil, err
	}

	b, err := encoder.EncodeExtended(o)
	if err != nil {
		return nil, fmt.Errorf("%w %v", ErrEncoding, err)
	}
	return b, nil
}

// register registers a encoder.Encoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func (er *encoderRegistry) register(enc encoder.Encoder) error {
	normalized := make(map[string]bool)

	exts := enc.Extensions()
	for _, ext := range exts {
		ext = strings.ToLower(ext)
		if _, found := normalized[ext]; found {
			return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
		}
		normalized[ext] = true
	}

	er.mutex.Lock()
	defer er.mutex.Unlock()

	for ext := range normalized {
		if _, ok := er.encoders[ext]; ok {
			return fmt.Errorf("extension '%s' %w", ext, ErrDuplicateFound)
		}
	}

	for ext := range normalized {
		er.encoders[ext] = enc
	}

	return nil
}

// deregister provides a mechanism for effectively removing the encoders from use
// for specific file types.
func (er *encoderRegistry) deregister(exts ...string) {
	er.mutex.Lock()
	defer er.mutex.Unlock()

	for _, ext := range exts {
		delete(er.encoders, strings.ToLower(ext))
	}

}
