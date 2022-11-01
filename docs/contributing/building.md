---
title: Building
weight: 20
---

You'll need a working `go` environment version >= 1.11 (official releases are built against `1.14.x`).
It's also expected that you have a Docker daemon for building images.

Clone the repository. The project uses [Go Modules](https://github.com/golang/go/wiki/Modules),
so if you explicitly define `$GOPATH` you should clone somewhere else.

Then, from the root directory:

```sh
make
```

This makes Docker images, and installs binaries to `$GOBIN` (if you define it) or `$(go env GOPATH)/bin`.

> **Note:** the default target architecture is amd64. If you would like
> to try to build Docker images and binaries for a different
> architecture you will have to set ARCH variable:
>
> ```sh
> $ make ARCH=<target_arch>
> ```

## Running tests

```sh
# Unit tests
make test

# End-to-end tests, acceptable Helm versions are v2,v3
make e2e HELM_VERSION=<version>

# Run specific end-to-end test
E2E_TESTS=./10_helm_chart.bats HELM_VERSION=v2 make e2e
```

For e2e tests to work on macOS, you may need to install some dependencies

```sh
brew install bash
brew install parallel
brew install coreutils
```