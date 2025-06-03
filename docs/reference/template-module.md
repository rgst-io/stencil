---
order: 3
---

# Template Module

A template module, sometimes referred to as a "template repository" is a module consumable by a stencil application, that contains a collection of go-templates.

## General Module Requirements

A module requires a name and a description currently that are set in the `manifest.yaml` described later on in this document.

The name of a module must be equal to the import path of the module. The import path of a module follows the same rules as Golang, where as the repository URL must equal the import path. For example, if the module is located at `https://github.com/stencil/example-module`, the import path must be `github.com/stencil/example-module`.

## Structure

A module structure typically looks like so:

- `templates/` - a directory that contains all of the go-templates that this module owns
- `manifest.yaml` - a manifest describing the arguments, dependencies and other metadata for this module
- `go.mod` - a go module file used for testing the module
- `**/.snapshots` - a directory used for snapshot testing files

### `templates/`

This directory is used for storing all of the go-templates that a module owns. By default a file that doesn't have a `.tpl` or `.nontpl` extension will be ignored by stencil.

- When a `.tpl` file is found, this file is written to the base of the execution directory of stencil, minus the `templates` directory and `.tpl` extension.
- Files with a `.nontpl` extension are treated as raw binary files and are written verbatim (not treated as templates) to the base of the execution directory of stencil, minus the `templates` directory and `.nontpl` extension.

For example, if a module had a template at `templates/helloWorld.tpl` it would by default be written to `./helloWorld`

This can be changed with the [`file.SetPath`](/funcs/file.SetPath) function as needed.

Templates can also call `file.Create` to create a new file within a loop. For more information see the [`file.Create` documentation](/funcs/file.Create)

#### Library Templates

Library templates special templates that are meant to only contain
functions callable by the current module. They cannot call `file`
methods as they do not ever generate files.

To create a library template, create a file with the `.library.tpl`
extension.

### `manifest.yaml`

The manifest.yaml file is arguably the most important file in a stencil module. This dictates the type of module, the arguments that the module accepts, and the dependencies that the module has.

The below list can also be found as Go struct at [pkg.go.dev](https://pkg.go.dev/go.rgst.io/stencil/v2/pkg/configuration#TemplateRepositoryManifest).

- `name` - The import path of the module
- `description` - A description of the module
- `modules` - a list of modules that this module depends on
  - `name` - import path of the module depended on
  - `version` - optional: A version to pin this module to.
- `postRunCommand` - An array of commands to run after the module is rendered. This is useful for running commands like `go mod tidy` or other build steps. The array entries have `name` and `command` keys:
  ```yaml
  - name: Prettier fix
    command: yarn run prettier:fix
  ```
- `dirReplacements` - a key:value mapping of template-able replacements for directory names, often used for languages like Java/Kotlin with directories named after the projects. These replacements can not rewrite directory structures, it only renames the leaf node directory name itself.
  - key: The directory name to replace
  - value: The template-able replacement name
  - example: This k:v pair will take the `com.projname` directory and replace it with the result of rendering the template (replacing it with the contents of the module's argument named "project-name"):
  ```yaml
  "src/main/kotlin/com.projname": '{{ stencil.Arg "project-name" }}'
  ```
  - If you need to rename nested directories, you need to know that stencil renders from the deepest directory up, so you must structure your renames like this:
  ```yaml
  "src/outer/inner": "bar"
  "src/outer": "foo"
  ```
  - This will result in the directory being named `src/foo/bar` after rendering -- the "outer" directory must match the actual pre-replace name in the filesystem.
- `arguments` - a map of arguments that this module accepts. A module cannot access an argument via `stencil.Arg` without first declaring it here.
  - `name` - the name of the argument
  - `description` - a description of the argument
  - `schema` - a JSON schema for the argument
  - `required` - whether or not the argument is required to be set
  - `default` - a default value for the argument, cannot be set when required is true
  - `from` - aliases this argument to another module's argument. Only
    supports one-level deep.
- `moduleHooks` - an optional map of a [module hook](#module-hooks)'s
  name to optional configuration.
  - `schema` - a JSON schema for the module hook, applies to each item
    being inserted into the module hook through `stencil.AddToModuleHook`.

#### Writing a JSON Schema

Arguments support JSON Schemas. The schema is used to validate the argument value. The schema is a JSON Schema [described here](https://json-schema.org/). This essentially boils down to two structures. For concrete types, like strings, numbers, and booleans, the schema is a simple object with a `type` key. For example:

```yaml
type: string
```

For more complex types, like objects, and arrays the schema is an object with properties or a list of properties. For example, objects:

```yaml
type: object
properties:
  name:
    type: string
```

For example, array of strings:

```yaml
type: array
items:
  type: string
```

#### Aliasing an argument with `from`

Aliasing an argument allows you to reference another argument from
within the module. For example, if you have an argument called
`description` and you want to alias it to another argument called from
the module `github.com/rgst-io/stencil-golang`, you can do so like so

```yaml
# your module
arguments:
  description:
    from: github.com/rgst-io/stencil-golang

# github.com/rgst-io/stencil-golang
arguments:
  description:
    schema:
      type: string
```

There's a few limitations with aliasing arguments:

- Aliasing an argument to another argument that is itself aliased is not allowed.
- When `from` is used, no other properties on the argument being aliased can be set.
- When aliasing to a module, that module _must_ be listed in the `modules` key of the module aliasing the argument.

## Module Hooks

Module hooks enable other modules to write to a section of a file in your module. This can be done with the [`stencil.GetModuleHook "name"`](/funcs/stencil.GetModuleHook) function. This returns a `[]any`, or for non-gophers a list of any type. You can process this with a `range` or in any other method you'd like to generate whatever you need for your DSL.

A module can write to a module hook with the [`stencil.AddToModuleHook "importPath" "hookName"`](/funcs/stencil.AddToModuleHook) function.

## Updating a Module

Modules, by default, are updated by default when running `stencil`. This is done by finding the latest Github release for a module and then using it. However, this may not be desired, so `stencil` can also be ran with the `--frozen-lockfile` command which will attempt to use the last ran versions again. An exception to this is major releases. Stencil will, by default, prompt the user for their permission to use the new version when a major version upgrade is detected. This will also display the release notes of that release to the user.

Module versions are stored in the `[]modules.version` keys in the `stencil.lock` file.

## Testing a Module

Testing a module can be done in a variety of different ways, but the officially supported way of testing a module is through the testing framework that's generated by the `stencil create module` command.

More documentation can be found on the documentation for that command, but in a nutshell the default and recommended testing method for modules is snapshot testing. Snapshot testing is done by rendering the files of a module to a directory, and then comparing the rendered files to the expected files overtime. This is supported by the [`stenciltest`](https://pkg.go.dev/github.com/rgst-io/stencil/pkg/stenciltest) go package.

Writing a test requires a valid `go.mod` file, as the tests are written in Go and to use the `stenciltest` package. To create a test, simply create a valid go test (e.g. `main_test.go`) and write a go test using the `stenciltest` package.

A simple example for rendering the template `helloWorld.tpl` to a file called `helloWorld` would look like so:

```
hello, world!
```

```go
package main

import (
	"testing"

  "github.com/rgst-io/stencil/pkg/stenciltest"
)

func TestGoMod(t \*testing.T) {
	// Create a renderer with the specified file being the file to test.
	//
	// More files may be provided if they are depended as variadic arguments
	// but their output will not be saved.
	st := stenciltest.New(t, "go.mod.tpl")

	// Define the arguments to pass to stencil
	st.Args(map[string]any{"org": "rgst-io"})

	// Run the test, persisting the snapshot to disk if it changed.
	// Default is set to false.
	st.Run(false)
}
```

You can run all tests by running `go test ./...` in the root of the repository:

```bash
$ go test ./...
...
ok      testing.com/templates  2.861s
```

### Testing a Module used in a Stencil Application

A `stencil.yaml` supports a `replacements` key that can be used to replace the source of a module with a different module. This is useful for testing a module that is used in a stencil application.

For example, if an application uses the `github.com/stencil/example-module` module and you want to develop on the `example-module`, a key in `replacements` can be added to point it to a different source URL or file path.

```yaml
replacements:
	# Replace it with a file path
	github.com/stencil/example-module: ../example-module

	# Replace it with a different URL
	github.com/stencil/example-module: github.com/myname/example-module
```

If you want to lock the dependency to a specific version currently replacements don't support setting the version, but instead you'd specify this in the `version` field of the module field of the `stencil.yaml`. For the example above:

```yaml
modules:
- name: github.com/stencil/example-module
	version: v1.0.0
```
