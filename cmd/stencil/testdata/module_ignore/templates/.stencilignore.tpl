{{- if stencil.Arg "skip" -}}
{{- file.Skip -}}
{{- else if stencil.Arg "delete" -}}
{{- file.Delete -}}
{{- end }}
go.mod
