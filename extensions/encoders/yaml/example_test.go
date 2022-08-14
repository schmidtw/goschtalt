// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"fmt"
	"os"

	"github.com/schmidtw/goschtalt"
	_ "github.com/schmidtw/goschtalt/extensions/encoders/yaml"
)

const text = `---
example:
	config:
		version: 1
		colors: [red, green, blue]`

func Example() {
	err := os.WriteFile("/tmp/example.yml", []byte(text), 0644)
	if err != nil {
		panic(err)
	}

	g, err := goschtalt.New(goschtalt.FileGroup(
		goschtalt.Group{
			FS:    os.DirFS("/tmp"),
			Paths: []string{"."}, // Look in '/tmp/.'
		}))
	if err != nil {
		panic(err)
	}

	err = g.Compile()
	if err != nil {
		panic(err)
	}

	var cfg struct {
		example struct {
			config struct {
				version int
				colors  []string
			}
		}
	}

	err = g.Unmarshal("", &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", cfg)

	// ...
}
