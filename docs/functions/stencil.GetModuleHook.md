---
order: 1060
---

# stencil.GetModuleHook

GetModuleHook returns a module block in the scope of this module.

This is incredibly useful for allowing other modules to write to files that your module owns. Think of them as extension points for your module. The value returned by this function is always an `[]any`, aka a list.

```
{{- /* This returns a []any */}}
{{ $hook := stencil.GetModuleHook "myModuleHook" }}
{{- range $hook }}
  {{ . }}
{{- end }}
```
