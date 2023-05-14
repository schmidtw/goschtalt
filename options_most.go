// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows && !android

package goschtalt

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const confDirName = "conf.d"

func stdCfgLayout(appName string, files []string) Option {
	var l stdLocations
	l.Populate(appName)
	return nonWinStdCfgLayout(appName, files, l)
}

type stdLocations struct {
	local fs.FS
	home  fs.FS
	etc   fs.FS
	root  fs.FS
}

func (s *stdLocations) Populate(name string) {
	s.local = os.DirFS(".")
	s.root = os.DirFS("/")
	s.etc = os.DirFS("/" + filepath.Join("etc", name))

	if home := os.Getenv("HOME"); home != "" {
		s.home = os.DirFS(filepath.Join(home, "."+name))
	}
}

func nonWinStdCfgLayout(appName string, files []string, paths stdLocations) Option {
	if appName == "" {
		return WithError(fmt.Errorf("%w: StdCfgLayout appName", ErrInvalidInput))
	}

	if len(files) > 0 {
		return AddJumbledHalt(paths.root, paths.local, files...)
	}

	single := appName + ".*"

	// The order of the options matters
	opts := []Option{
		AddFilesHalt(paths.local, single),
		AddTreeHalt(paths.local, confDirName),
	}

	if paths.home != nil {
		opts = append(opts,
			AddFilesHalt(paths.home, single),
			AddTreeHalt(paths.home, confDirName),
		)
	}

	opts = append(opts,
		AddFilesHalt(paths.etc, single),
		AddTreeHalt(paths.etc, confDirName),
	)

	return NamedOptions("StdCfgLayout", opts...)
}
