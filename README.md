# stencil

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/go.rgst.io/stencil)

A modern repository templating engine

**Note**: This has been forked from [getoutreach/stencil](https://github.com/getoutreach/stencil) and is currently
under construction.

## Development

### Prerequisites

**Note**: If you opt to not use `mise`, please install all dependencies
from `.tool-versions` manually.

- [mise](https://github.com/jdx/mise?tab=readme-ov-file#quickstart) - Ensure that you add the appropriate activations to your shell rc/profiles (details in the mise README)

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

Create a tag locally:

```bash
git tag -s vX.Y.Z
```

Push the tag:

```bash
git push origin refs/tags/vX.Y.Z
```

Wait for CI to build and publish the release (Github Actions).

## License

The original code, as of the fork, is licensed under the Apache 2.0
license which can be found (as it was) at
[LICENSE.original](LICENSE.original). This applies to all code created
before the commit [ea98384ec4b1031ba032cedad90df4bb0451cdce](https://go.rgst.io/stencil/commit/ea98384ec4b1031ba032cedad90df4bb0451cdce).

All code after that commit is licensed under AGPL 3.0. The license for
this can be found at [LICENSE](LICENSE).
