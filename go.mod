module aws-lambda-grafana-terraform-otlp

go 1.22

require (
	github.com/aws/aws-lambda-go v1.49.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.31.0
	go.opentelemetry.io/otel/metric v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0
)
