# Detailed Instruction: AWS Lambda Metrics to Grafana Cloud with Terraform

This instruction describes how to deploy an AWS Lambda function that sends custom OpenTelemetry metrics directly to Grafana Cloud using OTLP HTTP.

The project does not use EC2, Grafana Alloy, Prometheus server, or CloudWatch metric forwarding. Metrics are generated inside the Lambda function and sent directly to the Grafana Cloud OTLP endpoint.

## 1. Project goal

The goal of this project is to demonstrate a direct observability pipeline:

```
AWS Lambda
   ↓
OpenTelemetry SDK
   ↓
OTLP HTTP
   ↓
Grafana Cloud Metrics
   ↓
Grafana Dashboard
```

Terraform is used to create the AWS infrastructure automatically.

Terraform creates:

```
AWS IAM Role
AWS Lambda function
Lambda environment variables
Lambda runtime configuration
Lambda deployment package upload
```
The Go code is responsible for:

creating custom metrics
recording values during every Lambda invocation
sending metrics to Grafana Cloud through OTLP HTTP

## 2. Requirements

Before starting, you need:

```
AWS account
Grafana Cloud account
AWS CloudShell or local terminal
Terraform
Go
GitHub repository
```

The easiest way to run the project is through AWS CloudShell because it already has AWS authentication configured.

## 3. Grafana Cloud configuration

Open Grafana Cloud.

Go to:

```
Grafana Cloud Portal
↓
Your Stack
↓
OpenTelemetry
↓
Configure
```

Grafana Cloud will provide values similar to:

```
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-prod-REGION.grafana.net/otlp
OTEL_EXPORTER_OTLP_HEADERS=Authorization=Basic YOUR_TOKEN
```

These values are required for sending metrics from Lambda to Grafana Cloud.

Do not publish the real token in GitHub.

## 4. Create terraform.tfvars locally

The repository contains an example file:

```
terraform.tfvars.example
```

Create a real local file:

```
cp terraform.tfvars.example terraform.tfvars
```

Then edit terraform.tfvars:

```
aws_region = "eu-north-1"

otel_exporter_otlp_protocol = "http/protobuf"

otel_exporter_otlp_endpoint = "https://otlp-gateway-prod-REGION.grafana.net/otlp"

otel_exporter_otlp_headers = "Authorization=Basic YOUR_GRAFANA_CLOUD_OTLP_TOKEN"
```

## 5. Lambda metrics

The Lambda function sends custom metrics to Grafana Cloud.

The main metrics are:

Metric	Type	Description
```
lambda_invocations_total	Counter	Total number of Lambda invocations
lambda_success_total	Counter	Successful Lambda executions
lambda_errors_total	Counter	Simulated Lambda errors
lambda_cold_starts_total	Counter	Cold starts
lambda_processed_records_total	Counter	Simulated processed records
lambda_cache_hits_total	Counter	Simulated cache hits
lambda_cache_misses_total	Counter	Simulated cache misses
lambda_duration_ms	Histogram	Lambda execution duration
lambda_payload_size_bytes	Histogram	Incoming payload size
lambda_external_api_latency_ms	Histogram	Simulated external API latency
lambda_database_latency_ms	Histogram	Simulated database latency
lambda_random_business_value	Gauge	Simulated business value
lambda_memory_configured_mb	Gauge	Configured Lambda memory size
```

## 6. Build Lambda package

AWS Lambda with Go custom runtime requires the compiled binary to be named:
```
bootstrap
```

The build script creates this binary and packages it into:

```
lambda.zip
```
Run:
```
chmod +x build.sh
./build.sh
```
Expected result:
```
bootstrap
lambda.zip
```
The lambda.zip file is used by Terraform for deployment.

## 7. Terraform deployment

Initialize Terraform:
```
terraform init
```
Check what Terraform will create:
```
terraform plan
```
Apply the configuration:
```
terraform apply
```
Confirm with:
```
yes
```
Terraform will create:
```
IAM Role for Lambda
IAM policy attachment
AWS Lambda function
environment variables for OpenTelemetry and Grafana Cloud
```
## 8. Invoke Lambda

After deployment, the Lambda function can be invoked from AWS Console or AWS CloudShell.

Example AWS CLI command:

```
aws lambda invoke \
  --function-name lambda-grafana-terraform-direct \
  --cli-binary-format raw-in-base64-out \
  --payload '{"test":"lambda direct grafana metrics"}' \
  response.json \
  --region eu-north-1
```

To generate more metric data, run the Lambda several times:

```
for i in {1..20}; do
  aws lambda invoke \
    --function-name lambda-grafana-terraform-direct \
    --cli-binary-format raw-in-base64-out \
    --payload '{"test":"lambda direct grafana metrics"}' \
    response-$i.json \
    --region eu-north-1
done
```

## 9. Check metrics in Grafana

Open Grafana Cloud:

```
Grafana Cloud
↓
Launch Grafana
↓
Explore
↓
Prometheus / Grafana Cloud Metrics datasource
```

Use this query to show all Lambda metrics:

```
{__name__=~"lambda.*"}
```

10. PromQL queries for dashboards
Total invocations
```
sum(lambda_invocations_total)
```
Invocation rate
```
sum(rate(lambda_invocations_total[5m]))
```
Success count
```
sum(lambda_success_total)
```
Error count
```
sum(lambda_errors_total)
```
Error rate
```
sum(rate(lambda_errors_total[5m])) / sum(rate(lambda_invocations_total[5m])) * 100
```
Cold starts
```
sum(lambda_cold_starts_total)
```
Average duration
```
sum(rate(lambda_duration_ms_sum[5m])) / sum(rate(lambda_duration_ms_count[5m]))
```
P95 duration
```
histogram_quantile(0.95, sum(rate(lambda_duration_ms_bucket[5m])) by (le))
```
Average payload size
```
sum(rate(lambda_payload_size_bytes_sum[5m])) / sum(rate(lambda_payload_size_bytes_count[5m]))
```
External API latency
```
sum(rate(lambda_external_api_latency_ms_sum[5m])) / sum(rate(lambda_external_api_latency_ms_count[5m]))
```
Database latency
```
sum(rate(lambda_database_latency_ms_sum[5m])) / sum(rate(lambda_database_latency_ms_count[5m]))
```
Processed records rate
```
sum(rate(lambda_processed_records_total[5m]))
```
Cache hits
```
sum(rate(lambda_cache_hits_total[5m]))
```
Cache misses
```
sum(rate(lambda_cache_misses_total[5m]))
```
Cache hit ratio
```
sum(rate(lambda_cache_hits_total[5m])) / (sum(rate(lambda_cache_hits_total[5m])) + sum(rate(lambda_cache_misses_total[5m]))) * 100
```

The key result is a working serverless monitoring pipeline without EC2, without Grafana Alloy, and without CloudWatch metric forwarding:

AWS Lambda → OpenTelemetry OTLP HTTP → Grafana Cloud Metrics → Grafana Dashboard
