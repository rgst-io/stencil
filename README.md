# stencil

A modern living-template engine for evolving repositories.

Check out our [Documentation](https://stencil.rgst.io/) for more information!

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

### Building docs

If you have `mise` installed, you should have all the tooling you need for the docs engine. Run `mise run docsdev` from the root stencil directory to enter the watch-rebuild cycle to test your docs changes.

## License

The original code, as of the fork, is licensed under the Apache 2.0
license which can be found (as it was) at
[LICENSE.original](LICENSE.original). This applies to all code created
before the commit [ea98384ec4b1031ba032cedad90df4bb0451cdce](https://go.rgst.io/stencil/commit/ea98384ec4b1031ba032cedad90df4bb0451cdce).

All code after that commit is licensed under AGPL 3.0. The license for
this can be found at [LICENSE](LICENSE).
