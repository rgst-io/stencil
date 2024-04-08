---
order: 50
---

# file.RemoveAll

RemoveAll deletes all the files in the provided path. It has protections to limit paths to being under the directory being templated.

```
{{ file.RemoveAll "path" }}
```
