---
order: 1021
---

<!-- Generated by tools/docgen. DO NOT EDIT. -->

# stencil.GetModuleHook

GetModuleHook returns a module block in the scope of this module

This is incredibly useful for allowing other modules to write to files
that your module owns. Think of them as extension points for your
module. The value returned by this function is always a []any, aka a
list.

```go
{{- /* This returns a []any */}}
{{ $hook := stencil.GetModuleHook "myModuleHook" }}
{{- range $hook }}
  {{ . }}
{{- end }}
```
