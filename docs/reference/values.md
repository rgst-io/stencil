---
order: 21
---

# Template Values

Stencil passes different values via Go template's values context, for
authors of helm charts, this is like the `.Values` concept.

The fields available are:

- **`.Git`**: Information about the current Git repository (when present).
  - **`.Git.Ref`**: The current ref (e.g. `refs/heads/main`).
  - **`.Git.Commit`**: The current commit SHA.
  - **`.Git.Dirty`**: `true` when the working tree has uncommitted changes.
  - **`.Git.DefaultBranch`**: The repository's default branch (commonly `main`).

- **`.Runtime`**: Information about the tool/runtime generating files.
  - **`.Runtime.Generator`**: Name of the generator (usually `stencil`).
  - **`.Runtime.GeneratorVersion`**: The generator's version.
  - **`.Runtime.Modules`**: A slice of modules available during rendering.
    - Each module element has `.Name` and `.Version` ([*resolver.Version]) field.
    - Use the helper `.Runtime.Modules.ByName("modname")` to lookup a module by name.

- **`.Config`**: Manifest-derived configuration.
  - **`.Config.Name`**: The repository/project name from the manifest.

- **`.Module`**: The module currently being rendered.
  - **`.Module.Name`**: Module name.
  - **`.Module.Version`**: Module version ([*resolver.Version]).

- **`.Template`**: The template being rendered.
  - **`.Template.Name`**: Template name.

- **`.Data`**: Arbitrary data passed when a template is rendered via
  `stencil.Include` or `module.Call`. Present only for included
  templates and contains the value supplied by the caller.

[*resolver.Version]: https://pkg.go.dev/github.com/jaredallard/vcs/resolver#Version
