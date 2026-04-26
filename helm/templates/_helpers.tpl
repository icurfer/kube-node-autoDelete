{{/*
공통 chart 이름 (release/chart 기반)
*/}}
{{- define "kube-node-autodelete.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
fullname (RBAC, Deployment 등 리소스 이름)
*/}}
{{- define "kube-node-autodelete.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
ServiceAccount 이름
*/}}
{{- define "kube-node-autodelete.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "kube-node-autodelete.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
배포 대상 namespace
*/}}
{{- define "kube-node-autodelete.namespace" -}}
{{- default .Release.Namespace .Values.namespace.name -}}
{{- end -}}

{{/*
공통 라벨 (component 라벨은 각 템플릿에서 별도로 부여)
*/}}
{{- define "kube-node-autodelete.labels" -}}
app.kubernetes.io/name: {{ include "kube-node-autodelete.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: kube-node-autodelete
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end -}}

{{/*
selectorLabels (Deployment selector / matchLabels)
*/}}
{{- define "kube-node-autodelete.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-node-autodelete.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
