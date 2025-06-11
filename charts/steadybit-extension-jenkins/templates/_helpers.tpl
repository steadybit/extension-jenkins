{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "jenkins.secret.name" -}}
{{- default "steadybit-extension-jenkins" .Values.jenkins.existingSecret -}}
{{- end -}}
