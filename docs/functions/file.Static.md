---
order: 90
---

# file.Static

Static marks the current file as static.

Marking a file is equivalent to calling `file.Skip`, but instead `file.Skip` is only called if the file already exists. This is useful for files you want to generate but only once, with initial content. It's generally recommended that you do not do this as it limits your ability to change the file in the future.

```
{{ $_ := file.Static }}
```
