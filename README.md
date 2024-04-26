<div align="center">
  <a href="https://stencil.rgst.io">
    <img src=".github/images/stencil.png" alt="stencil" width="300" />
  </a>

  <p><em>A modern living-template engine for evolving repositories.</em></p>

  <a href="https://github.com/rgst-io/stencil/releases">
    <img alt="Latest Version" src="https://img.shields.io/github/v/release/rgst-io/stencil?style=for-the-badge">
  </a>
  <a href="https://github.com/rgst-io/stencil/blob/main/LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/rgst-io/stencil?style=for-the-badge">
  </a>
  <a href="https://github.com/rgst-io/stencil/actions/workflows/tests.yaml">
    <img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/rgst-io/stencil/tests.yaml?style=for-the-badge">
  </a>
  <a href="https://app.codecov.io/gh/rgst-io/stencil">
    <img alt="Codecov" src="https://img.shields.io/codecov/c/github/rgst-io/stencil?style=for-the-badge">
  </a>
  <hr />
</div>

Check out our [Documentation](https://stencil.rgst.io/) for more
information!

<!-- mise run update-readme-toc -->
<!-- toc -->
- [Install](#install)
  - [Homebrew](#homebrew)
  - [Binary](#binary)
  - [<code>go install</code>](#go-install)
- [Development](#development)
  - [Prerequisites](#prerequisites)
  - [Building](#building)
  - [Testing](#testing)
  - [Releasing](#releasing)
  - [Building docs](#building-docs)
- [License](#license)
<!-- /toc -->

## Install

### Homebrew

```bash
brew install rgst-io/tap/stencil
```

### Binary

Download the latest release from the [releases](/releases) page.

### `go install`

```bash
go install github.com/rgst-io/stencil@latest
```

## Development

### Prerequisites

**Note**: If you opt to not use `mise`, please install all dependencies
from `.tool-versions` manually.

- [mise](https://github.com/jdx/mise?tab=readme-ov-file#quickstart) -
Ensure that you add the appropriate activations to your shell
rc/profiles (details in the mise README)

Install the dependencies:

```bash
mise install
```

### Building

```bash
mise run build
```

### Testing

```bash
mise run test
```

### Releasing

Trigger a new releases by using the Github CLI, or through the
[Github Actions Workflow UI](https://github.com/rgst-io/stencil/actions/workflows/release.yaml).

```bash
gh workflow run release.yaml
```

Wait for CI to build and publish the release. You can watch the status
by running the following command:

```bash
gh workflow view release.yaml --web
```

### Building docs

If you have `mise` installed, you should have all the tooling you need
for the docs engine. Run `mise run docsdev` from the root stencil
directory to enter the watch-rebuild cycle to test your docs changes.

## License

The original code, as of the fork, is licensed under the Apache 2.0
license which can be found (as it was) at
[LICENSE.original](LICENSE.original). This applies to all code created
before the commit [ea98384ec4b1031ba032cedad90df4bb0451cdce](https://go.rgst.io/stencil/commit/ea98384ec4b1031ba032cedad90df4bb0451cdce).

All code after that commit is licensed under AGPL 3.0. The license for
this can be found at [LICENSE](LICENSE).
