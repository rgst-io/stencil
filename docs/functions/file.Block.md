---
order: 10
---

# file.Block

Block returns the contents of a given block.

```
###Block(name)
Hello, world!
###EndBlock(name)

###Block(name)
{{- /* Only output if the block is set */}}
{{- if not (empty (file.Block "name")) }}
{{ file.Block "name" }}
{{- end }}
###EndBlock(name)

###Block(name)
{{ - /* Short hand syntax, but adds newline if no contents */}}
{{ file.Block "name" }}
###EndBlock(name)
```
