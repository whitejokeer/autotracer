# 🛠️ AutoTracer – Setup OpenTelemetry (Prompt 00)

You are Claude, an expert Go developer specializing in OpenTelemetry instrumentation and observability best practices.

## Context
You're analyzing a Go repository to ensure it has proper OpenTelemetry bootstrapping. If the repository lacks a functional OpenTelemetry setup, you must create one following modern best practices.

## Prerequisites Check
First, verify if OpenTelemetry is already configured by checking for:
- Existing `internal/otel`, `pkg/otel`, or similar telemetry packages
- OpenTelemetry imports in `go.mod`
- Tracer initialization in main entry points

**If already configured**: Report the existing setup and skip implementation.

## Implementation Requirements

### 1. Dependencies
Add to `go.mod` (use latest stable versions):
```go
require (
    go.opentelemetry.io/otel v1.24.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
    go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.24.0
    go.opentelemetry.io/otel/sdk v1.24.0
    go.opentelemetry.io/otel/trace v1.24.0
    go.opentelemetry.io/contrib/instrumentation/runtime v0.49.0
)
```

### 2. Bootstrap Package
Create `internal/otel/otel.go` with:

```go
package otel

import (
    "context"
    "fmt"
    "os"
    "sync"
    "time"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
    "go.opentelemetry.io/otel/trace"
)

var (
    once     sync.Once
    tracer   trace.Tracer
    shutdown func(context.Context) error
)

// InitTracer initializes OpenTelemetry with the given service name.
// Returns a shutdown function that must be called on application exit.
func InitTracer(serviceName string) func(context.Context) error {
    once.Do(func() {
        ctx := context.Background()
        
        // Create resource
        res, err := resource.New(ctx,
            resource.WithAttributes(
                semconv.ServiceName(serviceName),
                semconv.ServiceVersion(getServiceVersion()),
            ),
            resource.WithTelemetrySDK(),
            resource.WithHost(),
            resource.WithProcess(),
        )
        if err != nil {
            panic(fmt.Errorf("failed to create resource: %w", err))
        }
        
        // Create exporter based on environment
        exporter, err := createExporter(ctx)
        if err != nil {
            panic(fmt.Errorf("failed to create exporter: %w", err))
        }
        
        // Create tracer provider
        tp := sdktrace.NewTracerProvider(
            sdktrace.WithBatcher(exporter),
            sdktrace.WithResource(res),
            sdktrace.WithSampler(sdktrace.AlwaysSample()),
        )
        
        // Set global providers
        otel.SetTracerProvider(tp)
        otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
            propagation.TraceContext{},
            propagation.Baggage{},
        ))
        
        tracer = tp.Tracer(serviceName)
        shutdown = func(ctx context.Context) error {
            return tp.Shutdown(ctx)
        }
    })
    
    return shutdown
}

// Tracer returns the global tracer instance
func Tracer() trace.Tracer {
    if tracer == nil {
        panic("tracer not initialized - call InitTracer first")
    }
    return tracer
}

func createExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
    endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if endpoint != "" {
        // OTLP exporter
        opts := []otlptracegrpc.Option{
            otlptracegrpc.WithEndpoint(endpoint),
            otlptracegrpc.WithTimeout(10 * time.Second),
        }
        
        // Add insecure option if not using TLS
        if os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true" {
            opts = append(opts, otlptracegrpc.WithInsecure())
        }
        
        client := otlptracegrpc.NewClient(opts...)
        return otlptrace.New(ctx, client)
    }
    
    // Fallback to stdout exporter
    return stdouttrace.New(
        stdouttrace.WithPrettyPrint(),
    )
}

func getServiceVersion() string {
    if version := os.Getenv("SERVICE_VERSION"); version != "" {
        return version
    }
    return "unknown"
}
```

### 3. Configuration File
Create `otel-config.yaml` (only if it doesn't exist):
```yaml
# OpenTelemetry Configuration
# 
# To use OTLP exporter, set environment variables:
#   OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
#   OTEL_EXPORTER_OTLP_INSECURE=true
#
# To use stdout exporter (default), leave OTEL_EXPORTER_OTLP_ENDPOINT unset

exporters:
  stdout:
    pretty_print: true
  
  otlp:
    endpoint: "localhost:4317"
    insecure: true

# Service configuration
service:
  # Set via SERVICE_VERSION env var or hardcode here
  version: "1.0.0"
```

### 4. Main Entry Point Integration
Locate the main entry point(s) and add initialization:

```go
import (
    "context"
    // ... other imports
    "{{MODULE_PATH}}/internal/otel"
)

func main() {
    // Initialize OpenTelemetry
    shutdown := otel.InitTracer("{{SERVICE_NAME}}")
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := shutdown(ctx); err != nil {
            log.Printf("Failed to shutdown tracer: %v", err)
        }
    }()
    
    // ... rest of main function
}
```

## Service Name Resolution
Determine service name in this order:
1. Binary name from `cmd/*/main.go` path
2. Module name from `go.mod`
3. Directory name as fallback

## Error Handling
- Use `panic` for initialization failures (fail fast)
- Log shutdown errors but don't fail the application
- Provide clear error messages with context

## Validation
After implementation, verify:
1. `go mod tidy` runs successfully
2. Code compiles without errors
3. Running with `OTEL_EXPORTER_OTLP_ENDPOINT` unset uses stdout
4. Tracer is accessible via `otel.Tracer()`

## Output Format
Provide:
1. Summary of changes made
2. Any existing setup found
3. Instructions for testing the setup
4. Environment variables needed for different exporters