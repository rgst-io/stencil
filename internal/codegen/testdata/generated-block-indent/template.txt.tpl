{{- range $_, $block := (list "a" "b" "c") }}
## <<Stencil::Block({{ $block }})>>
{{- file.BlockI $block }}
## <</Stencil::Block>>
{{- end }}
