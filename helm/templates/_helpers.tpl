{{/*
Expand the name of the chart.
*/}}
{{- define "reservio-slot-seeker.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "reservio-slot-seeker.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "reservio-slot-seeker.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "reservio-slot-seeker.labels" -}}
helm.sh/chart: {{ include "reservio-slot-seeker.chart" . }}
{{ include "reservio-slot-seeker.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "reservio-slot-seeker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "reservio-slot-seeker.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "reservio-slot-seeker.serviceAccountName" -}}
{{- $sa := default dict .Values.serviceAccount -}}
{{- $create := false -}}
{{- if hasKey $sa "create" -}}
{{- $create = index $sa "create" -}}
{{- end -}}
{{- $name := "" -}}
{{- if hasKey $sa "name" -}}
{{- $name = index $sa "name" -}}
{{- end -}}
{{- if $create -}}
{{- default (include "reservio-slot-seeker.fullname" .) $name }}
{{- else -}}
{{- default "default" $name }}
{{- end -}}
{{- end }}
