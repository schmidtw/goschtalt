<!--
SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
SPDX-License-Identifier: Apache-2.0
-->
# goschtalt
A simple configuration library that supports multiple files and formats.

[![Build Status](https://github.com/schmidtw/goschtalt/actions/workflows/ci.yml/badge.svg)](https://github.com/schmidtw/goschtalt/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/schmidtw/goschtalt/coverage.svg?branch=main)](http://codecov.io/github/schmidtw/goschtalt?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt)](https://goreportcard.com/report/github.com/schmidtw/goschtalt)
[![GitHub Release](https://img.shields.io/github/release/schmidtw/goschtalt.svg)](CHANGELOG.md)
[![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt)](https://pkg.go.dev/github.com/schmidtw/goschtalt)

## Goals & Themes

* Favor small, simple designs.
* Keep dependencies to a minimum.
* Favor user customization options over building everything in.
* Leverage go's new fs.FS interface for collecting files.

## API Stability

This package has not yet released, so APIs are subject to change for a bit longer.

## Extensions

These are just the extensions the goschtalt team maintains.  Others may be available
and it's fairly easy to write your own.

### Configuration Decoders

The decoders convert a file format into a useful object tree.  The meta.Object has
many convenience functions that make adding decoders pretty simple.  Generally,
the hardest part is determining where you are processing in the original file.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/decoders/cli)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/decoders/cli) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/decoders/cli)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/decoders/cli) | `extensions/decoders/cli` | A command line argument based configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/decoders/env)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/decoders/env) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/decoders/env)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/decoders/env) | `extensions/decoders/env` | An environment variable based configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/decoders/json)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/decoders/json) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/decoders/json)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/decoders/json) | `extensions/decoders/json` | A JSON configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/decoders/properties)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/decoders/properties) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/decoders/properties)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/decoders/properties) | `extensions/decoders/properties` | A properties configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/decoders/yaml)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/decoders/yaml) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/decoders/yaml)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/decoders/yaml) | `extensions/decoders/yaml` | A YAML/YML configuration decoder |


### Configuration Encoders

The encoders are used to output configuration into a file format.  Ideally you want
a format that accepts comments so it's easier see where the configurations originated
from.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/encoders/yaml)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/encoders/yaml) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/encoders/yaml)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/encoders/yaml) | `extensions/encoders/yaml` | A YAML/YML configuration encoder. |


### Opinionated CLI Integrations

The cli extensions are meant to cover basic use cases so you can focus on building
your app and not details like command line arguments for passing in the configuration
files you're interested in.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/schmidtw/goschtalt/extensions/cli/simple)](https://goreportcard.com/report/github.com/schmidtw/goschtalt/extensions/cli/simple) | [![GoDoc](https://pkg.go.dev/badge/github.com/schmidtw/goschtalt/extensions/cli/simple)](https://pkg.go.dev/github.com/schmidtw/goschtalt/extensions/cli/simple) | `extensions/cli/simple` | A fairly feature complete solution for simple servers that only need configuration. |

## Examples
