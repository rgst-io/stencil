global:
  deploymentEnvironment: prod
## <<Stencil::Block(version)>>
{{- if empty (trim (file.Block "version")) }}
  version: latest
{{- else }}
{{ file.Block "version"}}
{{- end }}
## <</Stencil::Block>>
somechart:
  somestuff: 4

## <<Stencil::Block(prod-{{ $ds.name }}-datasource)>>
breaksit
## <</Stencil::Block>>
