package main

import (
  "context"
  "os"
  "go.opentelemetry.io/otel"
  "go.opentelemetry.io/otel/attribute"
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
  "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
  "go.opentelemetry.io/otel/sdk/resource"
  sdktrace "go.opentelemetry.io/otel/sdk/trace"
  sdkmetric "go.opentelemetry.io/otel/sdk/metric"
  semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
  "google.golang.org/grpc"
)

func initOTel(ctx context.Context) (func(), error) {
  res, _ := resource.NewWithAttributes(
    semconv.SchemaURL,
    semconv.ServiceName("frontend"),
    semconv.ServiceVersion("1.0.0"),
    attribute.String("deployment.environment", "assessment"),
  )

  agentEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
  if agentEndpoint == "" {
    agentEndpoint = "localhost:4317"
  }

  traceExporter, _ := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint(agentEndpoint),
    otlptracegrpc.WithInsecure(),
    otlptracegrpc.WithDialOption(grpc.WithBlock()),
  )
  metricExporter, _ := otlpmetricgrpc.New(ctx,
    otlpmetricgrpc.WithEndpoint(agentEndpoint),
    otlpmetricgrpc.WithInsecure(),
  )

  tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(traceExporter),
    sdktrace.WithResource(res),
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
  )
  otel.SetTracerProvider(tp)

  mp := sdkmetric.NewMeterProvider(
    sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
    sdkmetric.WithResource(res),
  )
  otel.SetMeterProvider(mp)

  return func() {
    tp.Shutdown(ctx)
    mp.Shutdown(ctx)
  }, nil
}