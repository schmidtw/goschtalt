// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// Package goschtalt is a lightweight and flexible configuration registry that
// makes it easy to configure an application.
//
// Goschtalt is a fresh take on application configuration now that Go has
// improved filesystem abstraction, modules, and the Option pattern.  At its
// core, Goschtalt is a low dependency library that provides configuration
// values via a small and customizable API.  The configuration values can be
// merged using either the default semantics or specified on a parameter by
// parameter basis.
//
// # What problems is Goschtalt trying to solve?
//
//   - Merging of multiple configuration files and configuration sources allow
//     for better flexibility for deployment.
//   - Configuration is incrementally compiled allowing later configuration to
//     use earlier configuration.
//   - Clear explanation how the configuration was derived.
//   - User customizable via Options.
//
// # Features
//
//   - Users choose which configuration file decoders they want.
//   - Configuration fields may be labeled as 'secret' to enable secret redaction
//     during output of portions of the configuration tree.
//   - Configuration fields may instruct the merge process of how the new field
//     should merge with the existing field.  ('replace', 'keep', 'fail',
//     'append', 'prepend', 'clear')
//   - Configuration file groups include a reference to the specific io.fs, so
//     configuration may come from anything that implements that interface.
//   - Package defaults are set via goschtalt.DefaultOptions, but can be replaced
//     when invoking a new goschtalt.Config object.
//   - Default values are supported at runtime.
//   - Variable expansion in the configuration tree is supported for both
//     environment variables as well as custom values.
//   - No singleton objects.
//   - Low dependency count.
//
// # Where do I find configuration file encoders/decoders?
//
// The project contains several related packages that are isolated from the core
// goschtalt package to keep the dependencies low.  They are all maintained by
// the same group of people, but are not required to be used together, and are
// otherwise independent (different go modules).  They can be found here:
//
//   - https://github.com/goschtalt/env-decoder
//   - https://github.com/goschtalt/properties-decoder
//   - https://github.com/goschtalt/json-decoder
//   - https://github.com/goschtalt/yaml-encoder
//   - https://github.com/goschtalt/yaml-decoder
//
// # How do I decorate my configuration files to take full advantage of goschtalt?
//
// For most of the decoders you can specify instructions for goschtalt's handling
// of the data fields by annotating the key portion.  Here's a simple example in
// yaml:
//
//	foo:
//	  bar((prepend)):
//	    - 1
//	    - 2
//
// If this configuration data is merged with an existing configuration set:
//
//	foo:
//	  bar:
//	    - 3
//	    - 4
//
// The resulting configuration will be:
//
//	foo:
//	  bar:
//	    - 1
//	    - 2
//	    - 3
//	    - 4
//
// The commands available are consistent, but vary based on the type of the value
// being defined.
//
// All types (maps, arrays, values) support:
//   - replace - replaces any existing values encountered by this merge
//   - keep    - keeps the existing values encountered by this merge
//   - fail    - causes the merge to return an error and stop processing
//   - clear   - causes all of the existing configuration tree to be deleted
//   - secret  - this special command marks the field as secret
//
// Maps support the following instructions:
//   - splice  - merge the leaf nodes if possible instead of replacing the map entirely
//
// Arrays support the following instructions:
//   - append  - append this array to the existing array
//   - prepend - prepend this array to the existing array
//
// Default merging behaviors:
//   - maps   - splice when possible, replace if splicing isn't possible
//   - arrays - append
//   - values - replace
//
// An example showing using a secret:
//
//	foo:
//	  bar ((append secret)):
//	    - 3
//	    - 4
//
// The order of the instructions doesn't matter, nor does extra spaces around
// the instructions.  You may comma separate them, or you may just use a space.
// But you can only have one or two instructions (one MUST be secret if there are
// two.
//
// # A bit more on secrets.
//
// Secrets are primarily there so that if you want to output your configuration
// and everything is marked as secret correctly, you can get a redacted
// configuration file with minimal work.  It's also handy if you output your
// configuration values into a log so you don't accidentally leak your secrets.
//
// # How do I write my own configuration decoder?
//
// Examples of decoders exist in the extensions/decoders directory.  Of interest
// are the `env` decoder that provides an Option, and the `yaml` decoder that
// is simply a decoder.
//
// # What's with the name?
//
// In psychology, gestalt is a way of thinking about data via patterns and
// configuration.  It's a also somewhat common word.  gostalt is pretty good,
// except there were still several things that used it, including a go framework.
// This goschtalt project is the only response google returned as of Aug 12,
// 2022.
package goschtalt // import "github.com/goschtalt/goschtalt"
