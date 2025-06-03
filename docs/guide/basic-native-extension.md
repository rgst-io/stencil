---
order: 4
---

# Basic Native Extension

> [!NOTE]
> This quick start assumes you're familiar with stencil and module
> usage already. If you aren't, be sure to go through the
> [reference documentation](/reference/) or the other quick starts
> here before proceeding. You've been warned!
>
> Also, if you get stuck, take a look at [stencil-golang's native
> extension](https://github.com/rgst-io/stencil-golang/blob/main/internal/plugin/plugin.go).
> It's a great example which only fetches license information from the
> Github API!

## What is a Native Extension?

Native extensions are special module types that don't use go-templates
to integrate with stencil. Instead they expose template functions
written in another language that can be called by stencil templates.

## How to create a Native Extension

This quick start will focus on creating a Go native extension. While
other languages may work as well, there currently is no official
documentation or support for those languages (if you're interested in
another language please contribute it back!).

### Step 1: Create a Native Extension

Much like a module we're going to use the [`stencil create module`]
command to create a native extension.

```bash
mkdir helloworld; cd helloworld
stencil create module github.com/yourorg/helloworld
```

However, instead of using the `templates/` directory we're going to
create a `plugin/` directory.

```bash
rm -f templates; mkdir plugin
```

Now that we've created the `plugin/` directory we're going to created a
simple `plugin.go` file that'll implement the `Implementation` interface
and prints `helloWorld` when the `helloWorld` function is called.

```go
package main

import (
	"fmt"

	"github.com/rgst-io/stencil/pkg/extensions/apiv1"
)

// _ is a compile time assertion to ensure we implement
// the Implementation interface
var _ apiv1.Implementation = &TestPlugin{}

type TestPlugin struct{}

func (tp *TestPlugin) GetConfig() (*apiv1.Config, error) {
	return &apiv1.Config{}, nil
}

func (tp *TestPlugin) ExecuteTemplateFunction(t *apiv1.TemplateFunctionExec) (any, error) {
	if t.Name == "helloWorld" {
		return "helloWorld"
	}

  return nil, nil
}

func (tp *TestPlugin) GetTemplateFunctions() ([]*apiv1.TemplateFunction, error) {
	return []*apiv1.TemplateFunction{
		{
			Name: "helloWorld",
		},
	}, nil
}

func helloWorld() (any, error) {
	fmt.Println("👋 from the test plugin")
	return "hello from a plugin!", nil
}

func main() {
	err := apiv1.NewExtensionImplementation(&TestPlugin{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
```

Next, we'll need to build the plugin: `go build -o bin/plugin .`

### Step 2: Using in a Test Module

Let's create a `testmodule` to consume the native extension.

```bash
mkdir testmodule; cd testmodule
stencil create module github.com/yourorg/testmodule
```

Now let's create a `hello.txt.tpl` that consumes the `helloWorld` function.

```go
{{ extensions.Call "github.com/yourorg/helloworld" "helloWorld" }}
```

Ensure that the `manifest.yaml` for this module consumes the native extension:

```yaml
name: testmodule
modules:
  - name: github.com/yourorg/helloworld
```

### Step 3: Running the Test Module

Now, in order to test the native extension and the module consuming it we'll need to create a test application.

```bash
mkdir testapp; cd testapp
cat > stencil.yaml <<EOF
name: testapp
modules:
  - name: github.com/yourorg/testmodule
replacements:
  # Note: Replace these directories with their actual paths. This assumes they're
  # right behind our application in the directory tree.
  github.com/yourorg/helloworld: ../helloworld
  github.com/yourorg/testmodule: ../testmodule
EOF
```

Now, if we run `stencil` we should get a `hello.txt` file in our test application.

```bash
testapp ❯ stencil
INFO stencil v0.9.0
INFO Fetching dependencies
INFO  -> github.com/yourorg/helloworld local
INFO Loading native extensions
INFO Rendering templates
INFO Writing template(s) to disk
INFO  -> Created hello.txt

testapp ❯ cat hello.txt
helloWorld
```

Success! :tada:

## Reflection

To reflect, we've created a `hello.txt.tpl` file that calls the
`helloWorld` function in the native extension we implemented.
