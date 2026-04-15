{{- if stencil.Arg "skip" -}}
{{- file.Skip "argument 'skip' was set to true" -}}
{{- else if stencil.Arg "delete" -}}
{{- file.Delete -}}
{{- end }}
go.mod
