---
order: 1010
---

# stencil.AddToModuleHook

AddToModuleHook adds to a hook in another module.

This functions write to module hook owned by another module for it to operate on. These are not strongly typed so it's best practice to look at how the owning module uses it for now. Module hooks must always be written to with a list to ensure that they can always be written to multiple times.

```
{{- /* This writes to a module hook */}}
{{ stencil.AddToModuleHook "github.com/myorg/repo" "myModuleHook" (list "myData") }}
```
