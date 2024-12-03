# Hacking on image-builder-cli

Hacking on `image-builder` should be fun and is easy.

We have a unit tests some integration testing.

## Setup

To work on bootc-image-builder one needs a working Go [environment]. See
[go.mod](go.mod). 

To run the testsuite install the test dependencies as outlined in the
[github action](./.github/workflows/go.yml) under
"Install test dependencies".

## Code layout

The go source code of bib is under `./cmd/image-builder`. It uses the
[images](https://github.com/osbuild/images) library internally to
generate the imagess. Unit tests (and integration tests where it
makes sense) are expected to be part of a PR but we are happy to
help if those are missing from a PR.

## Build

Build by running:
```console
$ go build ./cmd/image-builder/
```

## Unit tests

Run the unit tests via:
```console
$ go test -short ./...
```

There are some integration tests that can be run via with:
```console
$ go test ./...
```
