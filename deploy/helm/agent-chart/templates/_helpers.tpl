{{- define "agent.fullname" -}}
{{ .Values.agent.name | default .Release.Name }}
{{- end -}}

{{- define "agent.labels" -}}
app.kubernetes.io/part-of: unboxd-platform
app.kubernetes.io/component: agent
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
unboxd.cloud/agent-kind: {{ .Values.agent.kind }}
{{- end -}}

{{- define "agent.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{ .Values.serviceAccount.name | default (include "agent.fullname" .) }}
{{- else -}}
{{ .Values.serviceAccount.name | default "default" }}
{{- end -}}
{{- end -}}

{{- define "agent.isKernel" -}}
{{- eq .Values.agent.kind "kernel" -}}
{{- end -}}
