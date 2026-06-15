{{- define "platform.labels" -}}
app.kubernetes.io/part-of: unboxd-platform
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
