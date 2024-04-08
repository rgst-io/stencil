---
order: 20
---

# file.Create

Create creates a new file that is rendered by the current template.

If the template has a single file with no contents this file replaces it.

```
{{- define "command" }}
package main

import "fmt"

func main() {
  fmt.Println("hello, world!")
}

{{- end }}

# Generate a "<commandName>.go" file for each command in .arguments.commands
{{- range $_, $commandName := (stencil.Arg "commands") }}
{{- file.Create (printf "cmd/%s.go" $commandName) 0600 now }}
{{- stencil.ApplyTemplate "command" | file.SetContents }}
{{- end }}
```
