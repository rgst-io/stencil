---
order: 4
---

# Native Extensions

A native extension follows the base rules of a [template module](template-module.html), but it may not contain any templates. This means a template module **CANNOT** contain a native extension.

The major difference between a module and a native extension is that instead of a `templates/` directory a `plugin/` directory is used to store the source code of the plugin.

## Creating a Native Extension

For details on how to create a native extension, check out the [getting started](/guide/basic-native-extension) documentation.

## Fetching a Native Extension

By default a native extension is fetched from Github releases using semantic-versioning. A binary must exist on the Github release matching [these specific formats](https://github.com/rgst-io/stencil/blob/main/internal/modules/nativeext/nativeext.go#L248).

## Testing a Native Extension

Currently stencil does not provide a testing framework for native extensions, but the recommend approach would be to use the snapshot testing framework provided by stencil or to build a system outside of stenciltest for this.

A native extension can be ran locally using the `replacements` key in an application's manifest (`stencil.yaml`), which is described in the [module documentation](template-module#testing-a-module-used-in-a-stencil-application). However, when doing this the native extension must write it's binary to `bin/plugin`.

## Debugging a Native Extension

The [`go-plugin`](https://github.com/hashicorp/go-plugin) library does not surface errors to stencil. Instead, it will raise the generic message `failed to create connection to extension: Unrecognized remote plugin message`. To determine a more precise error message, execute the native extension binary directly. The binary path can usually be found in bottom of the returned error. If not, the binary lives in the `bin/plugin` subdirectory of the native extension folder.

## How Native Extensions Work

Native extensions are implemented using the [go-plugin](https://github.com/hashicorp/go-plugin) using the [`net/rpc`](https://pkg.go.dev/net/rpc) transport layer. go-plugin, in simple terms, implements this by executing a plugin and negotiating with it to create a unix socket to communicate over with the native extension.

Once a connection has been established stencil communicates with the plugin over the following RPC methods (in Go this is the [`Implementation`](https://pkg.go.dev/github.com/rgst-io/stencil/pkg/extensions/apiv1#Implementation) interface).

1. `GetTemplateFunctions` RPC is called, which returns a list of declared functions this native extension implements. The format for this is defined as [`TemplateFunction`](https://pkg.go.dev/github.com/rgst-io/stencil/pkg/extensions/apiv1#TemplateFunction) in Go.
2. Stencil creates a wrapper go-template function to call the `ExecuteTemplateFunction` rpc based on the data returned by `GetTemplateFunctions`. The function is exposed at `importPath.function` via the `extensions.Call` method.
3. When the function is called via `extensions.Call "importPath.function"` the rpc `ExecuteTemplateFunction` is called with the arguments passed to the function. The format for this RPC is defined as [`TemplateFunctionExec`](https://pkg.go.dev/github.com/rgst-io/stencil/pkg/extensions/apiv1#TemplateFunctionExec) in Go.
4. The response from the RPC is returned directly to the go-template with no processing.

> [!TIP]
> While `ExecuteTemplateFunction`'s return value is a `any`, that is actually just a wrapper around the lower level implementation dubbed as [`implementationTransport`](https://github.com/rgst-io/stencil/blob/main/pkg/extensions/apiv1/transport.go#L25). The return value of that is `[]byte`, which is expected to be valid JSON to be decoded by stencil before being passed to the template that called it. This is done to ensure typed data is always able to be sent over.
