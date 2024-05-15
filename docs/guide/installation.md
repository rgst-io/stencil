---
order: 1
---

# Installation

Stencil is written in [Go](https://golang.org/) with support for multiple platforms. The latest release can be found at [Stencil Releases](https://github.com/rgst-io/stencil/releases).

Stencil currently provides pre-built binaries in amd64 and arm64 flavors for the following platforms:

- macOS (Darwin)
- Windows
- Linux

Stencil may also be compiled from source wherever the Go toolchain can run; e.g., on other operating systems such as DragonFly BSD, OpenBSD, Plan&nbsp;9, Solaris, and others. See <https://golang.org/doc/install/source> for the full set of supported combinations of target operating systems and compilation architectures.

## Quick Install

### Homebrew (macOS)

We have a brew formula for Stencil. It is recommended to install Stencil via Homebrew on macOS.

```bash
brew install rgst-io/tap/stencil
```

### Binary (Cross-platform)

Download the appropriate version for your platform from [Stencil Releases](https://github.com/rgst-io/stencil/releases). Once downloaded, the binary can be run from anywhere. You don't need to install it into a global location. This works well for shared hosts and other systems where you don't have a privileged account.

Ideally, you should install it somewhere in your `PATH` for easy use. `/usr/local/bin` is the most probable location.

### Source

Stencil is quite easy to build from source as well.

#### Prerequisite Tools

- Git
- [Mise](https://mise.jdx.dev/getting-started.html#quickstart)

#### Fetch from GitHub

```
git clone https://github.com/rgst-io/stencil/stencil.git
cd stencil
mise run build
cp ./bin/stencil "$(go env GOPATH)/bin/stencil"
```

## Upgrade Stencil

Upgrading Stencil is as easy as downloading and replacing the executable you’ve placed in your `PATH`.
