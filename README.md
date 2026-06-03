# AWS Lambda Metrics to Grafana Cloud with Terraform and OpenTelemetry

This project demonstrates how to send custom metrics from AWS Lambda directly to Grafana Cloud using OpenTelemetry OTLP HTTP.

The main feature of this project is that there is no EC2 instance, no Grafana Alloy, no Prometheus server, and no separate collector. The Lambda function sends metrics directly to the Grafana Cloud OTLP endpoint.

Terraform is used to create and configure the AWS infrastructure.

## Architecture

```
Terraform
   ↓
AWS Lambda + IAM Role
   ↓
Go Lambda function
   ↓
OpenTelemetry OTLP HTTP
   ↓
Grafana Cloud Metrics
   ↓
Grafana Dashboard
```
