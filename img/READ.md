# Grafana Screenshots

This folder contains screenshots from Grafana Cloud showing custom metrics sent directly from AWS Lambda through OpenTelemetry OTLP HTTP.

## Screenshots

### 01-total-invocations.png

Shows the total number of Lambda invocations during the selected time range.

### 02-success-total.png

Shows the total number of successful Lambda executions.

### 03-errors-total.png

Shows the total number of Lambda errors.

### 04-error-rate.png

Shows the percentage of failed Lambda executions compared to the total number of invocations.

### 05-cold-starts.png

Shows the number of Lambda cold starts.

### 06-configured-memory.png

Shows the amount of memory configured for the Lambda function. In this project, the Lambda function is configured with 128 MB of memory, and this value is sent to Grafana as a separate metric.

### 07-payload-size.png

Shows the average size of the input payload sent to the Lambda function during test invocations.

### 08-external-api-latency.png

Shows simulated external API latency.
