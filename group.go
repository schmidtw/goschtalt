// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/schmidtw/goschtalt/pkg/meta"
)

// Group is a filesystem and paths to examine for configuration files.
type Group struct {
	// FS is the filesystem to examine.
	FS iofs.FS

	// Paths are either exact files, or directories to examine for configuration.
	Paths []string

	// Recurse specifies if directories encoutered in the Paths should be examined
	// recursively or not.
	Recurse bool

	// FileReadTimeout provides a way to bound our file IO operations in the
	// event that hardware has issues or the underlying calls result in network
	// calls.
	FileReadTimeout time.Duration

	// Limiter is called to provide the client of goschtalt the opportunity to
	// rate limit the fetching process for the files.  The response function
	// is called when the fetching is complete.
	Limiter func() func()
}

func (group Group) enumerate(exts []string) ([]string, error) {
	var files []string

	for _, path := range group.Paths {
		// Make sure the paths are consistent across FS implementations with
		// go's documentation.  This prevents errors due to some FS accepting
		// invalid paths while others correctly reject them.
		if !iofs.ValidPath(path) {
			return nil, fmt.Errorf("path '%s' %w", path, iofs.ErrInvalid)
		}

		file, err := group.FS.Open(path)
		if err != nil {
			// Fail if the path is not found, otherwise, continue
			if errors.Is(err, iofs.ErrInvalid) || errors.Is(err, iofs.ErrNotExist) {
				return nil, err
			}
			continue
		}
		stat, err := file.Stat()
		if err != nil {
			_ = file.Close()
			continue
		}

		var found []string
		if !stat.IsDir() {
			found = []string{path}
		} else {
			if group.Recurse {
				_ = iofs.WalkDir(group.FS, path, func(file string, d iofs.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					}
					found = append(found, file)
					return nil
				})
			} else {
				found, _ = iofs.Glob(group.FS, path+"/*")
			}
		}
		_ = file.Close()

		files = append(files, matchExts(exts, found)...)
	}
	sort.Strings(files)

	return files, nil
}

func (group Group) collectAndDecode(decoders *decoderRegistry, file, keyDelimiter string) (meta.Object, error) {
	var m meta.Object

	buf := bytes.NewBuffer(nil)
	ext := strings.TrimPrefix(filepath.Ext(file), ".")

	f, err := group.FS.Open(file)
	if err != nil {
		return meta.Object{}, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return meta.Object{}, err
	}
	_, err = io.Copy(buf, f)
	_ = f.Close()
	if err != nil {
		return meta.Object{}, err
	}

	err = decoders.decode(ext, info.Name(), keyDelimiter, buf.Bytes(), &m)
	if err != nil {
		return meta.Object{}, err
	}

	return m, nil
}

func (group Group) walk(decoders *decoderRegistry, keyDelimiter string) ([]meta.Object, error) {
	exts := decoders.extensions()

	files, err := group.enumerate(exts)
	if err != nil {
		return nil, err
	}

	list := []meta.Object{}
	for _, file := range files {
		obj, err := group.collectAndDecode(decoders, keyDelimiter, file)
		if err != nil {
			return nil, err
		}
		list = append(list, obj)
	}
	return list, nil
}

func matchExts(exts, files []string) (list []string) {
	for _, file := range files {
		lc := strings.ToLower(file)
		for _, ext := range exts {
			if strings.HasSuffix(lc, "."+ext) {
				list = append(list, file)
			}
		}
	}

	return list
}
