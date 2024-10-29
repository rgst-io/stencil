global:
  deploymentEnvironment: prod
	version: xyz
  otherField: 1
local:
  deploymentEnvironment: prod
## <<Stencil::Block(version)>>
{{- if empty (trim (file.Block "version")) }}
  version: latest
{{- else }}
{{ file.Block "version"}}
{{- end }}
## <</Stencil::Block>>
  otherField: 2
somechart:
  somestuff: 4
somechart:
  somestuff: 4
