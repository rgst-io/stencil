---
order: 3
---

# Basic Stencil Module

> [!NOTE]
> This quick start assumes you're familiar with stencil usage already. If you aren't be sure to go through the [reference documentation](/reference/) or the [quick start](/guide/).

## Step 1: Create a module

Using the `stencil create` command we're able to quickly create a module, let's start with a simple hello world module.

```bash
mkdir helloworld; cd helloworld
stencil create module github.com/yourorg/helloworld
```

You'll notice that when running that command, stencil itself was ran. This is because `stencil create` uses the [`stencil-template-base`](https://github.com/getoutreach/stencil-template-base), the [`stencil-base`](https://github.com/getoutreach/stencil-base), and the [`stencil-circleci`](https://github.com/getoutreach/stencil-circleci) modules by default. This brings automatic CI and testing support to your module.

Let's briefly look at what it's created:

```bash
helloworld ❯ ls -alh
total 616
drwxr-xr-x   23 jaredallard  wheel   736B May  4 20:41 .
drwxr-xr-x    9 jaredallard  wheel   288B May  4 20:40 ..
drwxr-xr-x   32 jaredallard  wheel   1.0K May  4 20:41 .bootstrap
drwxr-xr-x    3 jaredallard  wheel    96B May  4 20:41 .circleci
-rw-r--r--    1 jaredallard  wheel   274B May  4 20:41 .editorconfig
drwxr-xr-x    4 jaredallard  wheel   128B May  4 20:41 .github
-rw-r--r--    1 jaredallard  wheel   795B May  4 20:41 .gitignore
-rw-r--r--    1 jaredallard  wheel   896B May  4 20:41 .releaserc.yaml
-rw-r--r--    1 jaredallard  wheel   417B May  4 20:41 CONTRIBUTING.md
-rw-r--r--    1 jaredallard  wheel    11K May  4 20:41 LICENSE.txt
-rw-r--r--    1 jaredallard  wheel   160B May  4 20:41 Makefile
-rw-r--r--    1 jaredallard  wheel   684B May  4 20:41 README.md
-rw-r--r--    1 jaredallard  wheel   118B May  4 20:41 bootstrap.lock
-rw-r--r--    1 jaredallard  wheel   5.5K May  4 20:41 go.mod
-rw-r--r--    1 jaredallard  wheel   101K May  4 20:41 go.sum
-rw-r--r--    1 jaredallard  wheel    74B May  4 20:41 manifest.yaml
drwxr-xr-x  375 jaredallard  wheel    12K May  4 20:41 node_modules
-rw-r--r--    1 jaredallard  wheel   474B May  4 20:41 package.json
drwxr-xr-x    5 jaredallard  wheel   160B May  4 20:41 scripts
-rw-r--r--    1 jaredallard  wheel   3.4K May  4 20:41 stencil.lock
-rw-r--r--    1 jaredallard  wheel   138B May  4 20:41 stencil.yaml
drwxr-xr-x    3 jaredallard  wheel    96B May  4 20:41 templates
-rw-r--r--    1 jaredallard  wheel   140K May  4 20:41 yarn.lock
```

A lot of files were created, the majority of these are niceties, like an automatic LICENSE (Apache-2.0), a README, and a CONTRIBUTING.md. Then there's also automatic .gitignore, circleci configuration for CI and a .releaserc for conventional commit powered releases. For more information for how to use these files, see the [template-base documentation](https://github.com/getoutreach/stencil-template-base).

The most important directory is the `templates/` directory, which will contain any templates we want to render.

## Step 2: Creating a Template

Let's create a template that creates a simple hello world message in Go. We'll start by creating a `hello.go.tpl` in the `templates/` directory.

```go
package main

func main() {
	fmt.Println("Hello, world!")
}
```

## Step 3: Consuming the Module in an Application

Now that we've done that, how do we test it locally without CI'ing up a full build? This is super easy with the `replacements` map in a [`stencil.yaml`](/reference/stencil.yaml).

Let's quickly create a test application:

```bash
mkdir testapp; cd testapp
cat > stencil.yaml <<EOF
name: testapp
modules:
	- name: github.com/yourorg/helloworld

replacements:
	# Replace ../helloworld with the path to your module.
	github.com/yourorg/helloworld: ../helloworld
EOF
```

Now if we run stencil on the test application, we should see the following:

```bash
testapp ❯ stencil
INFO[0000] stencil v1.14.2
INFO[0000] Fetching dependencies
INFO[0002]  -> github.com/yourorg/helloworld local
INFO[0002] Loading native extensions
INFO[0002] Rendering templates
INFO[0002] Writing template(s) to disk
INFO[0002]   -> Created hello.go
```

It looks like it created `hello.go` for us! Let's validate:

```bash
testapp ❯ cat hello.go
package main

func main() {
	fmt.Println("Hello, world!")
}
```

:tada: We have a hello world application!

## Step 4: Using a Block

Blocks are incredibly easy to use in Stencil.

> [!NOTE]
> In case you don't remember, blocks are areas in your generated code that you'd like to persist across runs.

Let's create our own block in the hello.go template from earlier:

```go
package main

func main() {
	fmt.Println("Hello, world!")

	// <<Stencil::Block(additionalMessage)>>
	{{- /* It's important to not indent the file.Block to prevent the indentation from being copied over and.. over again. */ }}
{{ file.Block "additionalMessage" }}
	// <</Stencil::Block>>
}
```

If we re-run stencil and look at `hello.go` we should see the following:

```go
package main

func main() {
	fmt.Println("Hello, world!")
	// <<Stencil::Block(additionalMessage)>>

	// <</Stencil::Block>>
}
```

If we add contents to the block and re-run stencil they'll be persisted across the run!

## Step 5: (Optional/Advanced) Creating Multiple Files

One of the powerful parts of stencil is the ability to create an arbitrary number of files with a single template. This is done with the [`file.Create`](/functions/file.Create) function. Let's create a `greeter.go.tpl` template in the `templates/` directory that'll create `<greeting>.go` based on the `greetings` argument.

```tpl
# This is important! We don't want to create a greeter.go file

{{- $_ := file.Skip "Generates multiple files" }}
{{- define "greeter" -}}
{{- $greeting := .greeting }}
package main

func main() {
fmt.Println("$greeting, world!")
}

{{- end -}}

{{- range $_, $greeting := stencil.Arg "greetings" }}

# Create a new $greeting.go file

{{- file.Create (printf "%s.go" $greeting) 0600 now }}

# We'll render the template greeter with $greeting as the values being passed to it

# Once we've done that we'll use the output to set the contents of the file we just created.

{{- stencil.ApplyTemplate "greeter" $greeting | file.SetContents }}
{{- end }}
```

> [!NOTE]
> Blocks are supported in multiple files! When `file.SetPath` is called the host is searched to see if a file already exists at that path, if it does it is searched to see if it contains any blocks, if it does they are loaded and accessible via `file.Block` as normal

Now let's modify the `manifest.yaml` to accept the argument `greetings`:

```yaml
arguments:
  greetings:
    description: A list of greetings to use
type: list
require: true
default: ["hello", "goodbye"]
```

If we now run stencil on the test application, we should see the following:

```bash
testapp ❯ stencil
INFO[0000] stencil v1.14.2
INFO[0000] Fetching dependencies
INFO[0002]  -> github.com/yourorg/helloworld local
INFO[0002] Loading native extensions
INFO[0002] Rendering templates
INFO[0002] Writing template(s) to disk
INFO[0002]   -> Created hello.go
INFO[0002]   -> Created goodbye.go
```

> [!NOTE] > `hello` and `goodbye` came from the default list of greetings that was set in the `manifest.yaml` file. Setting `arguments.greetings` on the test application and see it change!

If we look at the files, we should see the following:

```bash
testapp ❯ cat hello.go
package main

func main() {
	fmt.Println("hello, world!")
}

testapp ❯ cat goodbye.go
package main

func main() {
	fmt.Println("goodbye, world!")
}
```

## Reflection

We've created a module, used it in a test application via the `replacements` map in the `stencil.yaml` and used a block. Optionally we've also created multiple files with a template. This is just the beginning of what you can do with modules. Modules have a rich amount of [functions](/functions/) available to them. Check out the [reference](/reference/) for more information about modules and how to use them.
