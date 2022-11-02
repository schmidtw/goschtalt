// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// group is a filesystem and paths to examine for configuration files.
type group struct {
	// fs is the filesystem to examine.
	fs fs.FS

	// paths are either exact files, or directories to examine for configuration.
	paths []string

	// recurse specifies if directories encoutered in the paths should be examined
	// recursively or not.
	recurse bool
}

func (g group) enumerate(exts []string) ([]string, error) {
	var files []string

	// By default include everything in the base directory if nothing is specified.
	if len(g.paths) == 0 {
		g.paths = []string{"."}
	}

	for _, path := range g.paths {
		// Make sure the paths are consistent across FS implementations with
		// go's documentation.  This prevents errors due to some FS accepting
		// invalid paths while others correctly reject them.
		if !fs.ValidPath(path) {
			return nil, fmt.Errorf("path '%s' %w", path, fs.ErrInvalid)
		}

		file, err := g.fs.Open(path)
		if err != nil {
			// Fail if the path is not found, otherwise, continue
			if errors.Is(err, fs.ErrInvalid) || errors.Is(err, fs.ErrNotExist) {
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
			if g.recurse {
				_ = fs.WalkDir(g.fs, path, func(file string, d fs.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					}
					found = append(found, file)
					return nil
				})
			} else {
				found, _ = fs.Glob(g.fs, path+"/*")
			}
		}
		_ = file.Close()

		files = append(files, matchExts(exts, found)...)
	}
	sort.Strings(files)

	return files, nil
}

func (g group) collectAndDecode(decoders *codecRegistry[decoder.Decoder], file, keyDelimiter string) (meta.Object, error) {
	var m meta.Object
	var dec decoder.Decoder

	buf := bytes.NewBuffer(nil)
	ext := strings.TrimPrefix(filepath.Ext(file), ".")

	f, err := g.fs.Open(file)
	if err != nil {
		return m, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err == nil {
		_, err = io.Copy(buf, f)
		if err == nil {
			dec, err = decoders.find(ext)
		}
	}

	if err != nil {
		return m, err
	}

	ctx := decoder.Context{
		Filename:  info.Name(),
		Delimiter: keyDelimiter,
	}
	err = dec.Decode(ctx, buf.Bytes(), &m)
	if err != nil {
		err = fmt.Errorf("decoder error for extension '%s' processing file '%s' %w %v",
			ext, info.Name(), ErrDecoding, err)
	}

	return m, err
}

type fileObject struct {
	File string
	Obj  meta.Object
}

func (g group) walk(decoders *codecRegistry[decoder.Decoder], keyDelimiter string) ([]fileObject, error) {
	exts := decoders.extensions()

	files, err := g.enumerate(exts)
	if err != nil {
		return nil, err
	}

	var list []fileObject
	for _, file := range files {
		obj, err := g.collectAndDecode(decoders, file, keyDelimiter)
		if err != nil {
			return nil, err
		}
		list = append(list, fileObject{File: path.Base(file), Obj: obj})
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
