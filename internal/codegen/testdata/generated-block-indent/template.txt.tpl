{{- range $_, $block := (list "a" "b" "c") }}
###Block({{ $block }})
{{- file.BlockI $block }}
###EndBlock({{ $block }})
{{- end }}
