# SarosObserv

OpenTelemetry-Native Observability Platform

SarosObserv is an open-source, OpenTelemetry-native observability and application intelligence platform that provides unified collection, storage, visualization, correlation, anomaly detection, and evidence-based root-cause investigation for traces, metrics, and logs.

## Vision

Connect your application to SarosObserv through OpenTelemetry, and SarosObserv automatically discovers, visualizes, correlates, and analyzes your application.

**No manual dashboard creation. No manual service registration. Just connect OpenTelemetry.**

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- OpenTelemetry-instrumented application

### One-Command Deployment

```bash
docker compose up -d
```

This starts:
- **ClickHouse** (port 8123 HTTP, 9000 Native)
- **SarosObserv Backend** (ports 4317 OTLP/gRPC, 4318 OTLP/HTTP, 8080 API)
- **SarosObserv Frontend** (port 3000)

### Access the UI

Open your browser to: http://localhost:3000

## Investigate a request failure

Open **Overview** to see requests and errors by minute and open a recent failed
request. Its **Trace Detail** shows spans and up to 100 recent logs sharing the
trace ID. Request counts come from OpenTelemetry spans of kind `SERVER`, and an
error is a server span with OpenTelemetry status `ERROR`. Older spans stored
before span kind was captured still appear in the trace explorer.

With the stack running, send example checkout telemetry over OTLP/gRPC:

```bash
go run ./examples/go/investigation
```

Refresh Overview and select a failed `checkout-demo` request to see its
payment timeout log and attributes.

## Connect Your Application

Set the following environment variable in your application:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

### Language-Specific Examples

#### Node.js

```javascript
const { trace } = require('@opentelemetry/api');
const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { SimpleSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');

const provider = new NodeTracerProvider({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: 'my-service',
  }),
});

const exporter = new OTLPTraceExporter({
  url: 'http://localhost:4317',
});

provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
provider.register();
```

#### Python

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import OTLPSpanExporter
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource

resource = Resource.create({
    "service.name": "my-service"
})

provider = TracerProvider(resource=resource)
trace.set_tracer_provider(provider)

otlp_exporter = OTLPSpanExporter(endpoint="http://localhost:4317", insecure=True)
provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
```

#### Go

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func initTracer() error {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return err
    }

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String("my-service"),
        ),
    )
    if err != nil {
        return err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)

    return nil
}
```

## Architecture

```
Applications
     │
     │ OTLP (gRPC/HTTP)
     ▼
┌──────────────────────┐
│  SarosObserv         │
│  Ingestion Layer     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Telemetry Processing│
│  - Normalize         │
│  - Enrich            │
│  - Correlate         │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  ClickHouse Storage  │
│  - Traces            │
│  - Metrics           │
│  - Logs              │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  SarosObserv API    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  SarosObserv UI     │
└──────────────────────┘
```

## Development Roadmap

### Phase 1 — Telemetry Ingestion ✅
- OTLP receiver
- Trace ingestion
- ClickHouse storage

### Phase 2 — Trace Explorer (In Progress)
- Trace list
- Trace search
- Trace details
- Span tree
- Waterfall visualization

### Phase 3 — Service Discovery
- Automatic service discovery
- Service map
- Endpoint discovery
- Dependency graph

### Phase 4 — Metrics
- Metrics ingestion
- RED metrics (Rate, Errors, Duration)
- Metrics ↔ Traces correlation

### Phase 5 — Logs
- Log ingestion
- Full-text search
- Trace ↔ Logs correlation

### Phase 6 — Application Intelligence
- Endpoint health
- Service health
- Latency breakdown
- Error analysis

### Phase 7 — Anomaly Detection
- Baseline detection
- Latency anomalies
- Error anomalies

### Phase 8 — RCA Engine
- Incident detection
- Evidence collection
- Root cause analysis

## Project Structure

```
sarosobserv/
├── backend/              # Go backend
│   ├── cmd/
│   │   └── sarosobserv/  # Main application
│   ├── internal/
│   │   ├── ingestion/    # OTLP ingestion
│   │   ├── telemetry/    # Storage layer
│   │   ├── traces/       # Trace processing
│   │   ├── metrics/      # Metrics processing
│   │   ├── logs/         # Log processing
│   │   ├── services/     # Service discovery
│   │   ├── endpoints/    # Endpoint discovery
│   │   ├── dependencies/ # Dependency tracking
│   │   ├── correlation/  # Telemetry correlation
│   │   ├── analytics/    # Analytics engine
│   │   ├── incidents/    # Incident detection
│   │   ├── rca/          # Root cause analysis
│   │   └── api/          # REST API
│   └── pkg/
├── frontend/             # React TypeScript frontend
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── traces/
│   │   ├── services/
│   │   ├── endpoints/
│   │   ├── metrics/
│   │   ├── logs/
│   │   └── incidents/
├── storage/
│   └── clickhouse/       # ClickHouse schemas
├── deploy/
│   ├── docker/           # Dockerfiles
│   └── compose/          # Compose configurations
├── docs/                 # Documentation
├── examples/             # Example applications
│   ├── nodejs/
│   ├── python/
│   ├── go/
│   └── java/
└── docker-compose.yml    # Main deployment
```

## Ports

| Service | Port | Protocol |
|---------|------|----------|
| SarosObserv UI | 3000 | HTTP |
| SarosObserv API | 8080 | HTTP |
| OTLP/gRPC | 4317 | gRPC |
| OTLP/HTTP | 4318 | HTTP |
| ClickHouse HTTP | 8123 | HTTP |
| ClickHouse Native | 9000 | Native |

## Data Retention

Default retention policies:
- **Traces**: 7 days
- **Metrics**: 30 days
- **Logs**: 14 days

## Contributing

SarosObserv is in early development. Contributions are welcome!

## License

MIT License - See LICENSE file for details

## Philosophy

1. **OpenTelemetry First** - OTel is the universal telemetry interface
2. **Zero Configuration** - Users should not build dashboards before seeing value
3. **Correlation First** - Traces, metrics, and logs should not feel like separate products
4. **Evidence Before AI** - Collect and correlate evidence before generating explanations
5. **Investigation Over Visualization** - The goal is to understand what's happening, not just show pretty graphs

---

**OpenTelemetry in. Application intelligence out.**
