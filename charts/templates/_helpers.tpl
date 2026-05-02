{{/*
Expand the name of the chart.
*/}}
{{- define "court.name" -}}
{{- .Chart.Name -}}
{{- end }}

{{/*
Create a fully qualified release name.
Ensures uniqueness per namespace
*/}}
{{- define "court.fullname" -}}
{{- printf "%s" .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Common labels applied to all resources.
Follows Kubernetes recommended labels:
https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
*/}}
{{- define "court.labels" -}}
app.kubernetes.io/name: {{ include "court.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "court.selectorLabels" -}}
app.kubernetes.io/name: {{ include "court.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Component-specific labels
*/}}
{{- define "court.componentLabels" -}}
app.kubernetes.io/component: {{ .component }}
{{ include "court.labels" .context }}
{{- end }}

# ------------------------------------------------------------
# Resource names
# ------------------------------------------------------------

{{- define "court.officer.fullname" -}}
{{- printf "%s-officer" (include "court.fullname" .) -}}
{{- end }}

{{- define "court.archive.fullname" -}}
{{- printf "%s-archive" (include "court.fullname" .) -}}
{{- end }}

{{- define "court.docket.fullname" -}}
{{- printf "%s-docket" (include "court.fullname" .) -}}
{{- end }}

{{- define "court.court.fullname" -}}
{{- printf "%s-court" (include "court.fullname" .) -}}
{{- end }}

# ------------------------------------------------------------
# Connection helpers
# ------------------------------------------------------------

{{/*
PostgreSQL connection string
*/}}
{{- define "court.database.url" -}}
postgres://{{ .Values.archive.auth.user }}:{{ .Values.archive.auth.password }}@{{ include "court.archive.host" . }}:{{ .Values.archive.service.port }}/{{ .Values.archive.auth.database }}
{{- end -}}

{{/*
Archive (PostgreSQL) service hostname
*/}}
{{- define "court.archive.host" -}}
archive
{{- end }}

{{/*
Docket (Redis) service hostname
*/}}
{{- define "court.docket.host" -}}
docket
{{- end }}
