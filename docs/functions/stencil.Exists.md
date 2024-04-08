---
order: 1040
---

# stencil.Exists

Exists returns true if the file exists in the current directory.

```
{{- if stencil.Exists "myfile.txt" }}
{{ stencil.ReadFile "myfile.txt" }}
{{- end }}
```
