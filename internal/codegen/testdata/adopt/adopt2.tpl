global:
  deploymentEnvironment: prod
## <<Stencil::Block(version)>>
{{- if empty (trim (file.Block "version")) }}
  version: latest
{{- else }}
{{ file.Block "version"}}
{{- end }}
## <</Stencil::Block>>
  otherField: 1
local:
  deploymentEnvironment: prod
  version: abc
  otherField: 1
somechart:
  somestuff: 4
somechart:
  somestuff: 4
