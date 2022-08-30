// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"fmt"
	"os"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/cli/simple"
	_ "github.com/schmidtw/goschtalt/extensions/decoders/yaml"
)

const defaultConfigExt = `yml`
const defaultConfig = `---
# This is an example of documentation and default configuration.
# this defenition can live somewhere else, outside of this file
# or could be generated from a different file.  The choice is
# yours.

Example:
  # color can be one of [ red, green, blue ]
  color: blue # default to a popular color
  crayons ((secret)): box of 24
`

type all struct {
	Color   string
	Crayons string
}

func Example() {
	// Clear out arguments and make it easy to try out in goplayground.
	os.Args = []string{"exmaple"}

	program := simple.Program{
		Name: "example",
		Default: simple.DefaultConfig{
			Text: defaultConfig,
			Ext:  defaultConfigExt,
		},
		Licensing: "Apache-2.0",
		Validate: map[string]any{
			"example": all{},
		},
	}

	g, err := program.GetConfig()
	if g == nil { // We've been told to exit.
		if err != nil { // There was an error.
			panic(err)
		}
		return // No error, just gracefully exit.
	}

	// At this point you have processed the cli inputs, built a complete
	// configuration and are ready to go do things with it.
	//
	// Go forth and code...

	var s string
	s, _ = goschtalt.Fetch(g, "example.color", s)
	fmt.Println(s)

	// Output:
	// blue
}
