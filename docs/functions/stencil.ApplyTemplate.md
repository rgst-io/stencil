---
order: 1020
---

# stencil.ApplyTemplate

ApplyTemplate executes a template inside of the current module.

This function does not support rendering a template from another module.

```
{{- define "command"}}
package main

import "fmt"

func main() {
  fmt.Println("hello, world!")
}

{{- end }}

{{- stencil.ApplyTemplate "command" | file.SetContents }}
```
