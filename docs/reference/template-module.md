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

The important keys that a module has are listed below, but an exhaustive list can be found on the [pkg.go.dev](https://pkg.go.dev/github.com/rgst-io/stencil/pkg/configuration#TemplateRepositoryManifest)

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
  - `from` - aliases this argument to another module's argument. Only supports one-level deep.

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

Module hooks enable other modules to write to a section of a file in your module. This can be done with the [`stencil.GetModuleHook "name"`](/funcs/stencil.GetModuleHook) function. This returns a `[]interface{}`, or for non-gophers a list of any type. You can process this with a `range` or in any other method you'd like to generate whatever you need for your DSL.

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
	st.Args(map[string]interface{}{"org": "rgst-io"})

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

## Releasing a Module

Modules, when generated by the `stencil create` command, are configured to release differently based on the target merge branch.

- `main` - Creates a pre-release in Github, this is not automatically used or considered in stencil. These releases look like: `v0.0.0-rc.X`
- `release` - Creates a release in Github, this is considered automatically when stencil is ran.

### Conventional Commit

By default the versions bumps are made by using the commit messages. We recommend using squash and merge on your repository so that the title will be used for the commit, and thus the PR title is used instead. The rest of this document will assume this is the case.

A standard PR title should follow the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) format. This roughly translates to the format of `type(optionalScope): message`.

`feat: support multiple users`

`feat(users): add multiple user support`

Let’s look at each of the different sections of a conventional commit:

#### type

A type controls the what is released, or not released. Generally a type should be specific to the changes in the PR, but when in doubt you can always select one that has the releasing behavior you want.

##### **Major Release (vX.0.0)**

- **BREAKING CHANGE** - Breaks existing functionality. This will cause stencil to ask for a user's permission before updating
- **type!** - shorthand for BREAKING CHANGE, use any other type below with a ! at the end

Examples:

- `feat!(scope): break all the things`
- `feat!: break allllllll the things`

##### **Minor Release (v0.X.0)**

- `feat` - a feature, this should be something that adds to the project and is end user facing, this should not be a a breaking change

##### **Patch Release (v0.0.X)**

- `fix` - a fix to an existing feature in the project. Important: This should not be related to CI/CD, build, etc. See below for those
- `revert` - reverts a previous commit, e.g. an accidental breaking change
- `perf` - a performance modification, does not change existing functionality. Use feat if net-new functionality is added

##### **No Release**

- `refactor` - changes existing code, a catch-all for changes not related to performance
- `ci` - a modification related to the CI/CD system of the project. This does not trigger a release
- `build` - a modification related to the build of the system, e.g. docker file, scripts building it, etc. This does not trigger a release
- `docs` - a modification to the documentation of the project, e.g. README. This does not trigger a release
- `style` - a pure style change to existing source code (e.g. whitespace formatting)
- `test` - add missing tests or modify existing tests
- `chore` - a misc, catch-all, change that doesn’t modify source code. This does not trigger a release

#### scope

A scope is a useful way to separate changes, and identify them in a changelog. This is not required, and is loosely defined based on the project.

Examples:

- `feat(users): added multiple users modified files in a internal/reactor/users_controller.go`
- `fix(http): properly bind to config port modified the http server code in internal/reactor/httpservice.go`

### Creating a Release

Keeping in mind the release branches above, below are standard "SOPs" to release a module.

#### Creating a Release from `main` (Pre-release)

```bash
# Create the release
git checkout main; git pull
git checkout release; git pull
git merge main
git push

# Merge the merge commit back to main to ensure history is up-to-date.
git checkout main; git pull
git merge release
git push
```

#### Creating a One-Off Release (Hotfix)

```bash
# Create a cherry-picked commit
git checkout main; git pull
git checkout release; git pull
git cherry-pick <commit>
git push

# Keep main at the same history as release
git checkout main; git pull
git merge release
git push
```
