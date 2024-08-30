---
order: 2
---

# Base Project

This section will walk you through building a nearly-empty stenciled project and teach you how blocks work. While the sample project here is very Golang-centric, stencil can be used for any language!

## Step 0: Prerequisites

This tutorial requires the following things installed on your machine:

* Git - Used by the `stencil` CLI, this is always required.
* [mise] - This is not required usually, but the module we'll look at does
  require it.

## Step 1: Install Stencil

If you don't already have stencil installed, follow the [documentation to install it](/guide/installation.html).

## Step 2: Create a `stencil.yaml`

First start by creating a new directory for your application, this
should generally match whatever the name for your application is going
to be. We'll go with `helloworld` here.

```bash
mkdir helloworld
```

A [stencil.yaml](/reference/stencil.yaml) is integral to running
stencil. It defines the modules you're consuming and the arguments to
pass to them.

Start with creating a basic `stencil.yaml`:

```yaml
name: helloworld
arguments: {}
# Below is a list of modules to use for this application
# It should follow the following format:
# - name: <moduleImportPath>
#   version: "optional-version-to-pin-to"
modules: []
```

Now run `stencil`, you should have... nothing! That's expected because
we didn't define any modules yet.

```bash
helloworld ❯ stencil

INFO stencil 0.9.0
INFO Fetching dependencies
INFO Loading native extensions
INFO Rendering templates
INFO Writing template(s) to disk
INFO Running post-run command(s)

helloworld ❯ ls -alh
drwxr-xr-x 2 jaredallard jaredallard 4.0K Aug 29 19:51 .
drwxr-xr-x 8 jaredallard jaredallard 4.0K Aug 29 19:49 ..
-rw-r--r-- 1 jaredallard jaredallard   37 Aug 29 19:51 stencil.lock
-rw-r--r-- 1 jaredallard jaredallard   43 Aug 29 19:50 stencil.yaml
```

> [!NOTE]
> You'll notice there's a `stencil.lock` file here.
>
> ```yaml
> version: 0.9.0
> modules: []
> files: []
> ```
>
> This will keep track of what files were created by stencil and what
> created them, as well as the last ran version of your modules. This
> file is very important!

## Step 3: Import a Module

Now that we've created our first stencil application, you're going to
want to import a module! Let's import the
[`stencil-golang`](https://github.com/rgst-io/stencil-golang) module.
This module is for creating a Go service or CLI, but don't worry, you
don't need to know Go for this example!

First, let's take a look at parameters exposed through `stencil-golang`.
This can be done by looking at the `manifest.yaml` provided by it.

> [!NOTE] The full manifest can be found
> [here](https://github.com/rgst-io/stencil-golang/blob/e2ea9a1980f765f668d4a42d1f9108db777bf86d/manifest.yaml)

```yaml
name: github.com/rgst-io/stencil-golang
---
arguments:
  org:
    description: The Github organization to use for the project
    required: true

  library: <snip>
  copyrightHolder: <snip>
  license: <snip>
  commands: <snip>
```

We can see that the `org` argument is required and that it should be the
Github organization that our repository will live under. For now, since
we're not pushing this anywhere, it can be anything. So, let's go with
`rgst-io`.

```yaml
name: helloworld
arguments:
  org: rgst-io
modules:
  - name: github.com/rgst-io/stencil-golang
```

Now if we run stencil we'll see that we have some files!

```bash
helloworld ❯ stencil
INFO stencil 0.9.0
INFO Fetching dependencies
INFO  -> github.com/rgst-io/stencil-golang v1.0.0 (b81af6111ff23879d56faa735c428dc2e8ff45b5)
INFO Loading native extensions
INFO Rendering templates
INFO Writing template(s) to disk
INFO   -> Created LICENSE
INFO   -> Created .goreleaser.yaml
INFO   -> Created cmd/helloworld/helloworld.go
INFO   -> Created CONTRIBUTING.md
INFO   -> Created .github/workflows/release.yaml
INFO   -> Created .github/workflows/tests.yaml
INFO   -> Created .vscode/settings.json
INFO   -> Created .mise.toml
INFO   -> Created .github/scripts/get-next-version.sh
INFO   -> Created .github/settings.yml
INFO   -> Created .cliff.toml
INFO   -> Created .mise/tasks/changelog-release
INFO   -> Created package.json
INFO   -> Created .vscode/extensions.json
INFO   -> Created go.mod
INFO   -> Created .editorconfig
INFO   -> Created .gitignore
INFO   -> Created .vscode/common.code-snippets
INFO Running post-run command(s)
...

# Note: If you see an error about bun/prettier, simply run
# 'bun install' and re-run stencil. Sorry about that!

helloworld ❯ ls -alh
drwxr-xr-x 6 jaredallard jaredallard 4.0K Aug 29 19:58 .
drwxr-xr-x 8 jaredallard jaredallard 4.0K Aug 29 19:49 ..
-rw-r--r-- 1 jaredallard jaredallard 4.3K Aug 29 19:58 .cliff.toml
-rw-r--r-- 1 jaredallard jaredallard  185 Aug 29 19:58 .editorconfig
drwxr-xr-x 4 jaredallard jaredallard 4.0K Aug 29 19:58 .github
-rw-r--r-- 1 jaredallard jaredallard  583 Aug 29 19:58 .gitignore
-rw-r--r-- 1 jaredallard jaredallard 1.1K Aug 29 19:58 .goreleaser.yaml
drwxr-xr-x 3 jaredallard jaredallard 4.0K Aug 29 19:58 .mise
-rw-r--r-- 1 jaredallard jaredallard 1.3K Aug 29 19:58 .mise.toml
drwxr-xr-x 2 jaredallard jaredallard 4.0K Aug 29 19:58 .vscode
-rw-r--r-- 1 jaredallard jaredallard  217 Aug 29 19:58 CONTRIBUTING.md
-rw-r--r-- 1 jaredallard jaredallard  35K Aug 29 19:58 LICENSE
drwxr-xr-x 3 jaredallard jaredallard 4.0K Aug 29 19:58 cmd
-rw-r--r-- 1 jaredallard jaredallard   46 Aug 29 19:58 go.mod
-rw-r--r-- 1 jaredallard jaredallard  130 Aug 29 19:58 package.json
-rw-r--r-- 1 jaredallard jaredallard 2.4K Aug 29 19:58 stencil.lock
-rw-r--r-- 1 jaredallard jaredallard   96 Aug 29 19:58 stencil.yaml
```

Great, now we have all of these files! What stencil did was completely
generate this application for us as well as integrate with [mise] to set
up our toolchain.

## Step 4: Modifying a Block

One of the key features in stencil is the notion of "blocks". Modules
expose a block where they want developers to modify the code. Let's look
at the `stencil-golang` module to see what blocks are available.

Let's look at `.goreleaser.yaml`. This is a system for creating and
releasing Go binaries. Opening the file, we can see something like this:

```yaml
...
builds:
  - main: ./cmd/helloworld
    flags:
      - -trimpath
    ldflags:
      - -s
      - -w
      ## <<Stencil::Block(helloworldLdflags)>>

      ## <</Stencil::Block>>
...
```

A normal customization that Go application do is add `ldflags`, which
are linker-time modifications to variables. This is normally done with
something like a version string.

Here, we could add our own version string injection inside of the block.

```yaml
...
builds:
  - main: ./cmd/helloworld
    flags:
      - -trimpath
    ldflags:
      - -s
      - -w
      ## <<Stencil::Block(helloworldLdflags)>>
      - -X main.Version="myVersion"
      ## <</Stencil::Block>>
...
```

Now, if you re-run `stencil`, you'll see that it updated the file, but
did not change the contents. This is the "block" functionality that
makes up the core of `stencil`. This allows you to generate your files,
but most importantly _keep them updated_.

## Reflection

In all, we've created a `stencil.yaml`, added a module to it, ran
stencil, and then modified the contents within a block. That's it! We've
imported a base module and created some files, all without doing much of
anything on our side. Hopefully that shows you the power of stencil.

For more resources be sure to dive into [how to create a module](basic-module.md)
to get insight on how to create a module.

[mise]: https://mise.jdx.dev
