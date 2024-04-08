---
order: 40
---

# Native Extensions

Another important feature we added to stencil was support for [Native Modules](/reference/native-extensions.md). This allows you to add native code (we've used Golang, but other languages are viable as well since it uses a generic RPC interface for communication) to modules. We've used this for interfacing with other codegen systems like [goverter](https://github.com/jmattheis/goverter), running database migrations and extracting/codegenning data models from finished database schemas, and many other uses for when you need something more complicated than just living templates.

For an example of a basic native extension, check the [Guide](/guide/basic-native-extension) and it will probably be a little clearer.

## Next Steps

So that's the crash course. Does this sound interesting to you? Get started with our [Guide](/guide/installation)!
