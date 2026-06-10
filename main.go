package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Event map[string]interface{}

type Response struct {
	Message          string `json:"message"`
	Status           string `json:"status"`
	ProcessedRecords int64  `json:"processed_records"`
	PayloadSizeBytes int64  `json:"payload_size_bytes"`
}

var (
	meterProvider *sdkmetric.MeterProvider

	invocationCounter       metric.Int64Counter
	successCounter          metric.Int64Counter
	errorCounter            metric.Int64Counter
	coldStartCounter        metric.Int64Counter
	processedRecordsCounter metric.Int64Counter

	durationHistogram    metric.Float64Histogram
	payloadSizeHistogram metric.Int64Histogram

	memoryConfiguredGauge metric.Int64Gauge

	isColdStart = true
)

func initMeterProvider(ctx context.Context) error {
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("lambda5-grafana-terraform"),
			semconv.ServiceVersion("3.0.0"),
			attribute.String("deployment.environment", "student-demo"),
			attribute.String("cloud.provider", "aws"),
			attribute.String("cloud.platform", "aws_lambda"),
			attribute.String("faas.name", "lambda5"),
			attribute.String("project", "aws-lambda-grafana-terraform"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(5*time.Second),
	)

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(meterProvider)

	meter := otel.Meter("lambda5-direct-grafana")

	invocationCounter, err = meter.Int64Counter(
		"lambda5_invocations_total",
		metric.WithDescription("Total number of Lambda5 invocations"),
	)
	if err != nil {
		return err
	}

	successCounter, err = meter.Int64Counter(
		"lambda5_success_total",
		metric.WithDescription("Total number of successful Lambda5 executions"),
	)
	if err != nil {
		return err
	}

	errorCounter, err = meter.Int64Counter(
		"lambda5_errors_total",
		metric.WithDescription("Total number of real application errors in Lambda5"),
	)
	if err != nil {
		return err
	}

	coldStartCounter, err = meter.Int64Counter(
		"lambda5_cold_starts_total",
		metric.WithDescription("Total number of Lambda5 cold starts"),
	)
	if err != nil {
		return err
	}

	processedRecordsCounter, err = meter.Int64Counter(
		"lambda5_processed_records_total",
		metric.WithDescription("Total number of records processed by Lambda5"),
	)
	if err != nil {
		return err
	}

	durationHistogram, err = meter.Float64Histogram(
		"lambda5_duration_ms",
		metric.WithDescription("Lambda5 execution duration in milliseconds"),
	)
	if err != nil {
		return err
	}

	payloadSizeHistogram, err = meter.Int64Histogram(
		"lambda5_payload_size_bytes",
		metric.WithDescription("Size of incoming Lambda5 event payload in bytes"),
	)
	if err != nil {
		return err
	}

	memoryConfiguredGauge, err = meter.Int64Gauge(
		"lambda5_memory_configured_mb",
		metric.WithDescription("Configured Lambda5 memory size in MB"),
	)
	if err != nil {
		return err
	}

	return nil
}

func getMemorySizeMB() int64 {
	memoryValue := os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE")
	if memoryValue == "" {
		return 0
	}

	parsed, err := strconv.ParseInt(memoryValue, 10, 64)
	if err != nil {
		return 0
	}

	return parsed
}

func recordDuration(ctx context.Context, start time.Time, attrs metric.MeasurementOption) {
	durationMs := float64(time.Since(start).Milliseconds())
	durationHistogram.Record(ctx, durationMs, attrs)
}

func flushMetrics(ctx context.Context) {
	if err := meterProvider.ForceFlush(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush metrics: %v\n", err)
	}
}

func handler(ctx context.Context, event Event) (Response, error) {
	start := time.Now()

	if meterProvider == nil {
		if err := initMeterProvider(ctx); err != nil {
			return Response{
				Message: "OpenTelemetry initialization error",
				Status:  "error",
			}, err
		}
	}

	attrs := metric.WithAttributes(
		attribute.String("lambda_name", "lambda5"),
		attribute.String("project", "aws-lambda-grafana-terraform"),
		attribute.String("environment", "student-demo"),
	)

	invocationCounter.Add(ctx, 1, attrs)

	if isColdStart {
		coldStartCounter.Add(ctx, 1, attrs)
		isColdStart = false
	}

	payloadBytes, _ := json.Marshal(event)
	payloadSize := int64(len(payloadBytes))
	payloadSizeHistogram.Record(ctx, payloadSize, attrs)

	memoryConfiguredGauge.Record(ctx, getMemorySizeMB(), attrs)

	action, ok := event["action"].(string)
	if !ok || action != "process" {
		errorCounter.Add(ctx, 1, attrs)
		recordDuration(ctx, start, attrs)
		flushMetrics(ctx)

		return Response{
			Message:          "Invalid payload: field 'action' must be 'process'",
			Status:           "error",
			ProcessedRecords: 0,
			PayloadSizeBytes: payloadSize,
		}, fmt.Errorf("invalid payload: field 'action' must be 'process'")
	}

	recordsRaw, ok := event["records"].([]interface{})
	if !ok {
		errorCounter.Add(ctx, 1, attrs)
		recordDuration(ctx, start, attrs)
		flushMetrics(ctx)

		return Response{
			Message:          "Invalid payload: field 'records' must be an array",
			Status:           "error",
			ProcessedRecords: 0,
			PayloadSizeBytes: payloadSize,
		}, fmt.Errorf("invalid payload: field 'records' must be an array")
	}

	processedRecords := int64(len(recordsRaw))

	if processedRecords == 0 {
		errorCounter.Add(ctx, 1, attrs)
		recordDuration(ctx, start, attrs)
		flushMetrics(ctx)

		return Response{
			Message:          "Invalid payload: records array is empty",
			Status:           "error",
			ProcessedRecords: 0,
			PayloadSizeBytes: payloadSize,
		}, fmt.Errorf("invalid payload: records array is empty")
	}

	processedRecordsCounter.Add(ctx, processedRecords, attrs)
	successCounter.Add(ctx, 1, attrs)

	recordDuration(ctx, start, attrs)
	flushMetrics(ctx)

	return Response{
		Message:          "Lambda5 processed records and sent real metrics to Grafana Cloud",
		Status:           "ok",
		ProcessedRecords: processedRecords,
		PayloadSizeBytes: payloadSize,
	}, nil
}

func main() {
	lambda.Start(handler)
}
