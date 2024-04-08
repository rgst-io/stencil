---
order: 70
---

# file.SetPath

SetPath changes the path of the current file being rendered.

```
{{ $_ := file.SetPath "new/path/to/file.txt" }}
```

Note: The $\_ is required to ensure \<nil\> isn't outputted into the template\.
