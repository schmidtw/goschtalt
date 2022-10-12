// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	_ "github.com/schmidtw/goschtalt/extensions/decoders/yaml"
)

const filename = `example.yml`
const text = `---
example:
    version: 1
    colors: [red, green, blue]`

func getFS() fs.FS {
	mfs := memfs.New()
	if err := mfs.WriteFile(filename, []byte(text), 0755); err != nil {
		panic(err)
	}

	return mfs
}

func Example() {
	g, err := goschtalt.New(goschtalt.AddFileGroup(
		goschtalt.Group{
			FS:    getFS(),       // Normally, you use something like os.DirFS("/etc/program")
			Paths: []string{"."}, // Look in '.'
		}))
	if err != nil {
		panic(err)
	}

	err = g.Compile()
	if err != nil {
		panic(err)
	}

	var cfg struct {
		Example struct {
			Version int
			Colors  []string
		}
	}

	err = g.Unmarshal("", &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("example")
	fmt.Printf("    version = %d\n", cfg.Example.Version)
	fmt.Printf("    colors  = [ %s ]\n", strings.Join(cfg.Example.Colors, ", "))

	// Output:
	// example
	//     version = 1
	//     colors  = [ red, green, blue ]
}
