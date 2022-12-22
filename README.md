<!--
SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
SPDX-License-Identifier: Apache-2.0
-->
# goschtalt
A simple configuration library that supports multiple files and formats.

[![Build Status](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml/badge.svg)](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/goschtalt/goschtalt/coverage.svg?branch=main)](http://codecov.io/github/goschtalt/goschtalt?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt)](https://goreportcard.com/report/github.com/goschtalt/goschtalt)
[![GitHub Release](https://img.shields.io/github/release/goschtalt/goschtalt.svg)](https://github.com/goschtalt/goschtalt/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt)](https://pkg.go.dev/github.com/goschtalt/goschtalt)

## Goals & Themes

* Support multiple configuration files and sources.
* Enable tracing the origin of a configuration value.
* Favor user customization options over building everything in.
* Keep dependencies to a minimum.

## API Stability

This package has not yet released to 1.x yet, so APIs are subject to change for
a bit longer.

## Compilation of a Configuration

```mermaid
stateDiagram-v2
    direction TB

    state Gather_Inputs {
        direction TB
        AddBuffer() --> New()
        AddFile() --> New()
        AddValue(AsDefault()) --> New()
    }

    state Sequence {
        direction TB
        Defaults:Defaults by order added.
        Records:Records sorted using record label.
        ex:Expand instructions by order added.

        Defaults --> Records
        Records --> ex
    }

    state Compile {
        calc:Calculate configuration<br/>tree at thie point.
        eval:Apply all Expand()<br/>and ExpandEnv()<br/>to configuration<br/>tree in order.
        fetch:Call any user<br/>provided funcs with<br/>configuration tree.
        next:Next<br/>Configuration<br/>Tree Part
        merge:Merge the new<br/>configuration tree part<br/>with the current tree
        Empty:Empty<br/>Configuration<br/>Tree


        Empty --> calc
        calc --> eval:If a user func<br/>is provided
        calc --> next:If no user func<br/>is provided
        eval --> fetch
        fetch --> next
        next --> merge
        merge --> calc
    }

    state Expand {
        exp:Apply all Expand()<br/>and ExpandEnv()<br/>to configuration<br/>tree in order.
    }

    state Active {
        active:Active Configuration
        unmarshal:Unmarshal()
        active --> unmarshal
        unmarshal --> active
    }
    New() --> Sequence
    Sequence --> Compile
    Compile --> Expand
    Expand --> Active
    Active --> Sequence:With() or Compile() called<br/>resequences the lists and<br/>recalculates the <br/>configuration tree.
```

## Extensions

These are just the extensions the goschtalt team maintains.  Others may be available
and it's fairly easy to write your own.  Extensions have their own go.mod files
that independently track dependencies to keep dependencies only based on what
you need, not what could be used.

### Configuration Decoders

The decoders convert a file format into a useful object tree.  The meta.Object has
many convenience functions that make adding decoders pretty simple.  Generally,
the hardest part is determining where you are processing in the original file.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/env-decoder)](https://goreportcard.com/report/github.com/goschtalt/env-decoder) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/env-decoder)](https://pkg.go.dev/github.com/goschtalt/env-decoder) | n/a | An environment variable based configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/json-decoder)](https://goreportcard.com/report/github.com/goschtalt/json-decoder) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/json-decoder)](https://pkg.go.dev/github.com/goschtalt/json-decoder) | `.json` | A JSON configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/properties-decoder)](https://goreportcard.com/report/github.com/goschtalt/properties-decoder) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/properties-decoder)](https://pkg.go.dev/github.com/goschtalt/properties-decoder) | `.properties` | A properties configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/yaml-decoder)](https://goreportcard.com/report/github.com/goschtalt/yaml-decoder) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/yaml-decoder)](https://pkg.go.dev/github.com/goschtalt/yaml-decoder) | `.yaml`, `.yml` | A YAML/YML configuration decoder |


### Configuration Encoders

The encoders are used to output configuration into a file format.  Ideally you want
a format that accepts comments so it's easier see where the configurations originated
from.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/yaml-encoder)](https://goreportcard.com/report/github.com/goschtalt/yaml-encoder) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/yaml-encoder)](https://pkg.go.dev/github.com/goschtalt/yaml-encoder) | `.yaml`, `.yml` | A YAML/YML configuration encoder. |


## Dependencies

There are only two production dependencies in the core goschtalt code beyond the
go standard library.  The rest are testing dependencies.

Production dependencies:

* [github.com/mitchellh/hashstructure](https://github.com/mitchellh/hashstructure)
* [github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure)

## Examples

Coming soon.
