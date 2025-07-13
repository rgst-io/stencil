{{- range $_, $block := (list "a" "b" "c") }}
## <<Stencil::Block({{ $block }})>>
{{ file.Block $block }}
## <</Stencil::Block>>
{{- end }}
