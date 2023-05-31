// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package typical_test

import (
	"fmt"

	// Uncomment below use a real yaml decoder/encoder.  By not using these
	// in the goschtalt packages they aren't included in the dependencies.
	// This allows you to more easily control what decoders you want and also
	// allows easy to replacement in the future if a package becomes
	// unmaintained.
	//_ "github.com/goschtalt/yaml-decoder"
	//_ "github.com/goschtalt/yaml-encoder"

	"github.com/goschtalt/goschtalt"
	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/goschtalt/goschtalt/pkg/meta"
	_ "github.com/goschtalt/goschtalt/pkg/typical"
)

type Config struct {
	Name string
}

func Example() {
	gs, err := goschtalt.New(
		goschtalt.ConfigIs("two_words"),
		goschtalt.AddValue("built-in", goschtalt.Root,
			Config{
				Name: "app_name",
			},
			goschtalt.AsDefault(),
		),
		onlyNeededForExample(),
	)

	if err != nil {
		panic(err)
	}

	c, err := goschtalt.Unmarshal[Config](gs, goschtalt.Root)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Name: %q\n", c.Name)

	// Output:
	// Name: "app_name"
}

// -- Normally the code below isn't needed. ------------------------------------

type fake struct{}

func (*fake) EncodeExtended(meta.Object) ([]byte, error)         { return nil, nil }
func (*fake) Encode(v any) ([]byte, error)                       { return nil, nil }
func (*fake) Decode(decoder.Context, []byte, *meta.Object) error { return nil }
func (*fake) Extensions() []string                               { return []string{"yml", "yaml"} }

func onlyNeededForExample() goschtalt.Option {
	// Normally these are provided by yaml-decoder and yaml-encoder.
	return goschtalt.Options(
		goschtalt.WithDecoder(&fake{}),
		goschtalt.WithEncoder(&fake{}),
	)
}
