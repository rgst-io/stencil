{{ .Module.Version }} {{ (.Runtime.Modules.ByName "testing").Version }} {{ (index .Runtime.Modules 0).Version }} {{ .Template.Name }}
