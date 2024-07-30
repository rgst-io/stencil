---
order: 2
---

# Base Project

This section will walk you through building a nearly-empty stenciled project and teach you how blocks work. While the sample project here is very Golang-centric, stencil can be used for any language!

> [!NOTE]
> This quick start uses `macOS` in the examples. For instructions about how to install Stencil on other operating systems, see [install](installation.md).
> It is required to have [Git installed](https://git-scm.com/downloads) to run this tutorial.

## Step 1: Install Stencil

If you don't already have stencil installed, follow the [documentation to install it](/guide/installation.html).

## Step 2: Create a `stencil.yaml`

First start by creating a new directory for your application, this should generally match whatever the name for your application is going to be. We'll go with `helloworld` here.

```bash
mkdir helloworld
```

A [stencil.yaml](/reference/stencil.yaml) is integral to running stencil. It defines the modules you're consuming and the arguments to pass to them.

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

Now run `stencil`, you should have... nothing! That's expected because we didn't define any modules yet.

```bash
helloworld ❯ stencil

INFO[0000] stencil v1.14.2
INFO[0002] Fetching dependencies
INFO[0002] Loading native extensions
INFO[0002] Rendering templates
INFO[0002] Writing template(s) to disk
INFO[0002] Running post-run command(s)

helloworld ❯ ls -alh
drwxr-xr-x 4 jaredallard wheel 128B May 4 20:16 .
drwxr-xr-x 9 jaredallard wheel 288B May 4 20:16 ..
-rw-r--r-- 1 jaredallard wheel 213B May 4 20:16 stencil.yaml
-rw-r--r-- 1 jaredallard wheel 78B May 4 20:16 stencil.lock
```

> [!NOTE]
> You'll notice there's a `stencil.lock` file here.
>
> ```yaml
> version: 0.8.0
> modules: []
> files: []
> ```
>
> This will keep track of what files were created by stencil and what created them, as well as the last ran version of your modules. This file is very important!

## Step 3: Import a Module

Now that we've created our first stencil application, you're going to want to import a module! Let's import the [`stencil-base`](https://github.com/getoutreach/stencil-base) module. stencil-base includes a bunch of scripts and other building blocks for a Golang project. Let's take a look at it's `manifest.yaml` to see what arguments are required.

```yaml
name: github.com/getoutreach/stencil-base
---
arguments:
  description:
    required: true
    type: string
    description: The purpose of this repository
```

We can see that `description` is a required argument, so let's add it! Modify the `stencil.yaml` to set `arguments.description` to `"My awesome service!"`

```yaml
name: helloworld
arguments: {}
description: "My awesome service!"
modules:
  - name: github.com/getoutreach/stencil-base
```

Now if we run stencil we'll see that we have some files!

```bash
helloworld ❯ stencil
INFO[0000] stencil v1.14.2
INFO[0000] Fetching dependencies
INFO[0001]  -> github.com/getoutreach/stencil-base v0.2.0
INFO[0001] Loading native extensions
INFO[0001] Rendering templates
INFO[0001] Writing template(s) to disk
INFO[0001]   -> Created .editorconfig
INFO[0001]   -> Created .github/CODEOWNERS
INFO[0001]   -> Created .github/pull_request_template.md
INFO[0001]   -> Created .gitignore
INFO[0001]   -> Created .releaserc.yaml
INFO[0001]   -> Created CONTRIBUTING.md
INFO[0001]   -> Skipped LICENSE
INFO[0001]   -> Created README.md
INFO[0001]   -> Skipped helpers
INFO[0001]   -> Created package.json
INFO[0001]   -> Created scripts/devbase.sh
INFO[0001]   -> Created scripts/shell-wrapper.sh
INFO[0001] Running post-run command(s)
...

helloworld ❯ ls -alh
drwxr-xr-x   4 jaredallard  wheel   128B May  4 20:16 .
drwxr-xr-x   9 jaredallard  wheel   288B May  4 20:16 ..
-rw-r--r--   1 jaredallard  wheel   274B May  4 20:26 .editorconfig
drwxr-xr-x   4 jaredallard  wheel   128B May  4 20:26 .github
-rw-r--r--   1 jaredallard  wheel   795B May  4 20:26 .gitignore
-rw-r--r--   1 jaredallard  wheel   857B May  4 20:26 .releaserc.yaml
-rw-r--r--   1 jaredallard  wheel   417B May  4 20:26 CONTRIBUTING.md
-rw-r--r--   1 jaredallard  wheel   703B May  4 20:26 README.md
-rw-r--r--   1 jaredallard  wheel   474B May  4 20:26 package.json
drwxr-xr-x   4 jaredallard  wheel   128B May  4 20:26 scripts
-rw-r--r--   1 jaredallard  wheel   2.1K May  4 20:26 stencil.lock
-rw-r--r--   1 jaredallard  wheel   118B May  4 20:26 stencil.yaml
```

## Step 4: Modifying a Block

One of the key features in stencil is the notion of "blocks". Modules expose a block where they want developers to modify the code. Let's look at the `stencil-base` module to see what blocks are available.

In `README.md` we can see a basic block called "overview":

```md
# helloworld

...

## High-level Overview

<!--- <<Stencil::Block(overview)>> -->

<!--- <</Stencil::Block>> -->
```

Let’s add some content in two places. One inside the block, one outside:

```md
...

## High-level Overview

hello, world!

<!--- <<Stencil::Block(overview)>> -->

hello, world!

<!--- <</Stencil::Block>> -->
```

If we re-run stencil, notice how the contents of `README.md` have changed:

```md
...

## High-level Overview

<!--- <<Stencil::Block(overview)>> -->

hello, world!

<!--- <</Stencil::Block>> -->
```

The contents of `README.md` have changed, but the contents within the block have not. This is the power of blocks, modules are able to change the content _around_ a user's content without affecting the user's content. This can be taken even further if a template decides to parse the code within a block at runtime, for example using the ast package to rewrite go code.

## Reflection

In all, we've created a `stencil.yaml`, added a module to it, ran stencil, and then modified the contents within a block. That's it! We've imported a base module and created some files, all without doing much of anything on our side. Hopefully that shows you the power of stencil.

For more resources be sure to dive into [how to create a module](basic-module.md) to get insight on how to create a module.
