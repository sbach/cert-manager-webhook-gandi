{{/* vim: set filetype=mustache: */}}

{{/*
Note: we truncate at 63 chars because some Kubernetes name fields are
limited to this (by the DNS naming specification).
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-webhook-gandi.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified application name.
Note: The release name will be used if it contains the chart name.
*/}}
{{- define "cert-manager-webhook-gandi.fullname" -}}
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

{{- define "cert-manager-webhook-gandi.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "cert-manager-webhook-gandi.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "cert-manager-webhook-gandi.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-gandi.rootCAIssuer" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-gandi.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-gandi.rootCACertificate" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-gandi.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-gandi.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "cert-manager-webhook-gandi.fullname" .) }}
{{- end -}}
