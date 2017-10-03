Use the code {{.Code}} to verify your phone number {{if .AppName}}on {{.AppName}}{{end}}
{{if .URLToken -}}
    or visit {{.URLToken}}
{{- end}}