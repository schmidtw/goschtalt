// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// Provides a simple way to get all the options you typically expect to
// be set out of the box in goschtalt.
//
// # Why not include this automatically?
//
//	1. Due to separate packages for the adapters that depends on goschtalt,
//	   goschtalt can't also depend upon the adapters, so a 3rd package is needed.
//	2. If you don't want these values, no problem, simply don't include this package.
//	3. Testing is easier with no default values.
//
// # What is included?
//
//   - [goschtalt.AutoCompile]() - no need to compile explicitly
//   - [goschtalt.ExpandEnv]() - environment variables in the form of ${var} are replaced
//   - [goschtalt.IncludeOrigins]() - the origin of the configuration is included
//   - [goschtalt.FormatAs]("yml") - the output format used for the configuration is yml
//   - [goschtalt.Strictness](SUBSET) - the configuration can be a subset, but not contain extra things
//   - [adapter.BoolUnmarshal]() / [adapter.MarshalBool]() - convert bool to/from string
//   - [adapter.DurationUnmarshal]() / [adapter.MarshalDuration]() - convert [time.Duration] to/from string
//   - [adapter.FloatUnmarshal]() / [adapter.MarshalFloat]() - convert float32/float64 to/from string
//   - [adapter.IntUnmarshal]() / [adapter.MarshalInt]() - convert int/int8/int16/int32/int64 to/from string
//   - [adapter.TextUnmarshal]() / [adapter.MarshalText]() - convert objects that implement UnmarshalText/TextMarshal to/from string
//   - [adapter.TimeUnmarshal]() / [adapter.MarshalTime]() - convert [time.Time] to/from string (in RFC3339 form)
//   - [adapter.UintUnmarshal]() / [adapter.MarshalUint]() - convert uint/uint8/uint16/uint32/uint64 to/from string
//
// # Usage
//
// Add the following lines to the import list & everything is automatically ready
// to go.
//
//	import (
//		_ "github.com/goschtalt/goschtalt/pkg/typical"
//		_ "github.com/goschtalt/yaml-decoder"
//		_ "github.com/goschtalt/yaml-encoder"
//	)
package typical

import (
	"time"

	"github.com/goschtalt/goschtalt"
	"github.com/goschtalt/goschtalt/pkg/adapter"
)

func init() {
	goschtalt.DefaultOptions = append(goschtalt.DefaultOptions, typical())
}

// typical provides a fairly good set of options for using goschtalt.
//
// You must provide yaml encoders/decoders.  The following imports work great,
// but you are also free to choose whatever variations you wish.
func typical() goschtalt.Option {
	return goschtalt.Options(
		goschtalt.AutoCompile(),
		goschtalt.ExpandEnv(),
		goschtalt.DefaultMarshalOptions(
			goschtalt.IncludeOrigins(),
			goschtalt.FormatAs("yml"),
		),
		goschtalt.DefaultUnmarshalOptions(
			adapter.BoolUnmarshal(),
			adapter.DurationUnmarshal(),
			adapter.FloatUnmarshal(),
			adapter.IntUnmarshal(),
			adapter.TextUnmarshal(adapter.AllButTime),
			adapter.TimeUnmarshal(time.RFC3339),
			adapter.UintUnmarshal(),
			goschtalt.Strictness(goschtalt.SUBSET),
		),
		goschtalt.DefaultValueOptions(
			adapter.MarshalBool(),
			adapter.MarshalDuration(),
			adapter.MarshalFloat(),
			adapter.MarshalInt(),
			adapter.MarshalText(adapter.AllButTime),
			adapter.MarshalTime(time.RFC3339),
			adapter.MarshalUint(),
		),
		goschtalt.HintDecoder("yaml", "https://github.com/goschtalt/yaml-decoder", "yml", "yaml"),
		goschtalt.HintEncoder("yaml", "https://github.com/goschtalt/yaml-encoder", "yml", "yaml"),
	)
}
