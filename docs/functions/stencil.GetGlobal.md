---
order: 1050
---

# stencil.GetGlobal

GetGlobal retrieves a global variable set by SetGlobal\. The data returned from this function is unstructured so by averse to panics -- look at where it was set to ensure you're dealing with the proper type of data that you think it is.

```
{{- /* This retrieves a global from the current context of the template module repository */}}
{{ $isGeorgeCool := stencil.GetGlobal "IsGeorgeCool" }}
```
