---
order: 2
---

# Stencil Manifest

Every stenciled project starts with a `stencil.yaml` file, which is a basic manifest describing the project, the modules it uses, and arguments, which are used by the referenced modules to render out the different templates and/or execute native code.

## What are the fields in a `stencil.yaml`

- `name`: The name of the application
- `arguments`: The arguments to pass to the modules. This is a map of key value pairs.
- `modules`: The modules to use. This is a list of objects containing a `name` and a, optionally, `version` field to use of this module.
- `replacements`: A key/value of importPath to replace with another source. This is useful for replacing modules with a different version or local testing. Source should be a valid URL, import path, or file path on disk.
