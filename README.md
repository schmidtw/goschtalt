<!--
SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
SPDX-License-Identifier: Apache-2.0
-->
# goschtalt

[![Build Status](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml/badge.svg)](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/goschtalt/goschtalt/coverage.svg?branch=main)](http://codecov.io/github/goschtalt/goschtalt?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt)](https://goreportcard.com/report/github.com/goschtalt/goschtalt)
[![GitHub Release](https://img.shields.io/github/release/goschtalt/goschtalt.svg)](https://github.com/goschtalt/goschtalt/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt)](https://pkg.go.dev/github.com/goschtalt/goschtalt)

A customizable configuration library that supports multiple files and formats.

## Goals & Themes

* Make support for multiple configuration files and sources easy to use and extend.
* Simplify tracing the origin of a configuration value.
* Favor user customization options over building everything in, keeping dependencies
  to a minimum.
* Embrace patterns that make using other powerful libraries ([kong](https://github.com/alecthomas/kong), [fx](https://github.com/uber-go/fx), etc) simple.

## API Stability

This package has not yet released to 1.x yet, so APIs are subject to change for
a bit longer.

After v1 is released, [SemVer](http://semver.org/) will be followed.

## Installation

```shell
go get github.com/goschtalt/goschtalt
```

## Extensions

Instead of trying to build everything in, goschtalt tries to only build in what
is absolutely necessary and favor extensions.  This enables a diversity of
ecosystem while not bloating your code with a bunch of dependencies.

The following are popular goschtalt extensions:

* Name nomenclature converter https://github.com/goschtalt/casemapper
* Environment variable decoder https://github.com/goschtalt/env-decoder
* JSON file type decoder https://github.com/goschtalt/json-decoder
* Properties file type Decoder https://github.com/goschtalt/properties-decoder
* YAML file type decoder https://github.com/goschtalt/yaml-decoder
* YAML file type encoder https://github.com/goschtalt/yaml-encoder

## Examples

Coming soon.

## Dependencies

There are only two production dependencies in the core goschtalt code beyond the
go standard library.  The rest are testing dependencies.

Production dependencies:

* [github.com/mitchellh/hashstructure](https://github.com/mitchellh/hashstructure)
* [github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure)

## Compilation of a Configuration

This is more detailed overview of how the configuration is compiled.

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
        calc:Calculate configuration<br/>tree at this point.
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
