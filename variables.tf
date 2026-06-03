variable "aws_region" {
  description = "AWS region where Lambda will be created"
  type        = string
  default     = "eu-north-1"
}

variable "otel_exporter_otlp_protocol" {
  description = "OpenTelemetry protocol for Grafana Cloud OTLP endpoint"
  type        = string
  default     = "http/protobuf"
}

variable "otel_exporter_otlp_endpoint" {
  description = "Grafana Cloud OTLP endpoint"
  type        = string
  sensitive   = true
}

variable "otel_exporter_otlp_headers" {
  description = "Grafana Cloud OTLP authorization header"
  type        = string
  sensitive   = true
}
