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
//   - Multiple configuration files allow for better flexibility for deployment.
//   - Low dependency count.
//   - User customizable via Options.
//   - Able to collect configuration files from anywhere via the io.fs interface.
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
//   - Defaults are set via goschtalt.DefaultOptions, but can be replace when
//     invoking a new goschtalt object.
//   - No singleton objects.
//   - Only 1 non-standard library dependency on mitchellh/mapstructure, which
//     has no dependencies outside the standard library.
//
// # Where do I find configuration file decoders?
//
// TODO Currently these are a work in progress.
//
// # How do I decorate my configuration files to take advantage of goschtalt?
//
// TODO This documentation needs to be written.
//
// # How do I write my own configuration decoder?
//
// TODO This documentation needs to be written.
//
// # What's with the name?
//
// In psychology, gestalt is a way of thinking about data via patterns and
// configuration.  It's a also somewhat common word.  gostalt is pretty good,
// except there were still several things that used it, including a go framework.
// This goschtalt project is the only response google returned as of Aug 12,
// 2022.
package goschtalt // import "github.com/schmidtw/goschtalt"
