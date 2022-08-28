// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package simple_test

import (
	"fmt"
	"os"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/cli/simple"
	"github.com/schmidtw/goschtalt/pkg/decoder"

	// Uncomment the next line to use the yaml decoder automatically.
	//_ "github.com/schmidtw/goschtalt/pkg/decoder/yaml"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// This is a fake decoder that allows the example to work without the need
// to bring any external decoders in.
type fake struct{}

func (f fake) Decode(_ decoder.Context, _ []byte, _ *meta.Object) error { return nil }
func (f fake) Extensions() []string                                     { return []string{"yml"} }

func alterArgs() {
	os.Args = []string{"example", "--kvp", "Example.color", "blue", "--kvp", "Muppet", "fozzie"}

	// Uncomment the next line to output all the configuration values.
	// os.Args = append(os.Args, "-s")
}

func Example() {
	alterArgs() // This makes the function act like the command line.

	// Create a default configuration with all the options listed, documented,
	// and set to the default value, so the end user can ask the program for the
	// document.  It's ok to leave large sections commented out, too.
	defaults := simple.DefaultConfig{
		Text: `---
Example:
  # color can be one of [ red, green, blue ]
  color: blue # a popular color
`,
		Ext: "yml",
	}

	g, err := simple.GetConfig("example", defaults, goschtalt.DecoderRegister(fake{}))
	if g == nil { // We've been told to exit.
		if err != nil { // There was an error.
			panic(err)
		}
		// No error, just gracefully exit.
		return
	}

	// Now let's us the configuration...

	var s string
	s, _ = goschtalt.Fetch(g, "example.color", s)
	fmt.Println(s)

	s, _ = goschtalt.Fetch(g, "muppet", s)
	fmt.Println(s)

	// Output:
	// blue
	// fozzie
}
