package ingestion

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/sarosobserv/sarosobserv/backend/internal/telemetry"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

// OTLPServer handles OTLP ingestion.
type OTLPServer struct {
	storage  telemetry.Storage
	grpcAddr string
	httpAddr string
}

// NewOTLPServer creates a new OTLP server.
func NewOTLPServer(
	storage telemetry.Storage,
	grpcAddr string,
	httpAddr string,
) *OTLPServer {
	return &OTLPServer{
		storage:  storage,
		grpcAddr: grpcAddr,
		httpAddr: httpAddr,
	}
}

// traceService implements the OTLP gRPC trace service.
type traceService struct {
	ptraceotlp.UnimplementedGRPCServer
	server *OTLPServer
}

// Export receives OTLP traces over gRPC.
func (t *traceService) Export(
	ctx context.Context,
	req ptraceotlp.ExportRequest,
) (ptraceotlp.ExportResponse, error) {

	traces := req.Traces()

	if err := t.server.ProcessTraces(ctx, traces); err != nil {
		return ptraceotlp.NewExportResponse(), err
	}

	return ptraceotlp.NewExportResponse(), nil
}

// StartGRPC starts the OTLP gRPC server.
func (s *OTLPServer) StartGRPC(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return fmt.Errorf(
			"failed to listen on %s: %w",
			s.grpcAddr,
			err,
		)
	}

	server := grpc.NewServer()

	// Register the standard OTLP TraceService.
	ptraceotlp.RegisterGRPCServer(
		server,
		&traceService{
			server: s,
		},
	)

	log.Printf(
		"OTLP gRPC server listening on %s",
		s.grpcAddr,
	)

	go func() {
		<-ctx.Done()

		log.Println("Stopping gRPC server...")

		server.GracefulStop()
	}()

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf(
				"OTLP gRPC server stopped: %v",
				err,
			)
		}
	}()

	return nil
}

// StartHTTP starts the OTLP HTTP listener.
//
// OTLP HTTP is not implemented yet. The listener is kept here
// so the configured port can be reserved without pretending
// that OTLP HTTP ingestion is supported.
func (s *OTLPServer) StartHTTP(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.httpAddr)
	if err != nil {
		return fmt.Errorf(
			"failed to listen on %s: %w",
			s.httpAddr,
			err,
		)
	}

	log.Printf(
		"OTLP HTTP server listening on %s (full implementation pending)",
		s.httpAddr,
	)

	<-ctx.Done()

	return listener.Close()
}

// ProcessTraces processes trace data from OTLP.
func (s *OTLPServer) ProcessTraces(
	ctx context.Context,
	traces ptrace.Traces,
) error {

	spans := processPTrace(traces)

	for _, span := range spans {
		if err := s.storage.StoreSpan(ctx, span); err != nil {
			return fmt.Errorf(
				"failed to store span %s: %w",
				span.SpanID,
				err,
			)
		}
	}

	log.Printf(
		"Processed %d spans",
		len(spans),
	)

	return nil
}

// processPTrace converts OpenTelemetry pdata traces
// into SarosObserv SpanData objects.
func processPTrace(
	traces ptrace.Traces,
) []telemetry.SpanData {

	var spans []telemetry.SpanData

	resourceSpans := traces.ResourceSpans()

	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)

		serviceName := "unknown"

		resource := rs.Resource()

		// Extract service.name from resource attributes.
		serviceNameAttr, ok := resource.Attributes().Get(
			"service.name",
		)

		if ok && serviceNameAttr.Type() == pcommon.ValueTypeStr {
			serviceName = serviceNameAttr.AsString()
		}

		log.Printf(
			"Resolved service name: %q",
			serviceName,
		)

		scopeSpans := rs.ScopeSpans()

		for j := 0; j < scopeSpans.Len(); j++ {
			ss := scopeSpans.At(j)

			otelSpans := ss.Spans()

			for k := 0; k < otelSpans.Len(); k++ {
				span := otelSpans.At(k)

				startTime := span.StartTimestamp().AsTime()
				endTime := span.EndTimestamp().AsTime()

				durationNs := uint64(0)

				if endTime.After(startTime) {
					durationNs = uint64(
						endTime.Sub(startTime).Nanoseconds(),
					)
				}

				spanData := telemetry.SpanData{
					SpanID:       span.SpanID().String(),
					TraceID:      span.TraceID().String(),
					ParentSpanID: span.ParentSpanID().String(),

					ServiceName:   serviceName,
					OperationName: span.Name(),

					StartTime:  startTime,
					EndTime:    endTime,
					DurationNs: durationNs,

					Status: "ok",

					Attributes: convertPAttributes(
						span.Attributes(),
					),

					Events: convertPEvents(
						span.Events(),
					),
				}

				// OpenTelemetry status code:
				// 0 = Unset
				// 1 = OK
				// 2 = Error
				if span.Status().Code() != 0 {
					spanData.Status =
						span.Status().Code().String()

					spanData.StatusMessage =
						span.Status().Message()
				}

				spans = append(
					spans,
					spanData,
				)
			}
		}
	}

	return spans
}

// convertPAttributes converts OpenTelemetry attributes
// into a simple string map.
func convertPAttributes(
	attrs pcommon.Map,
) map[string]string {

	result := make(map[string]string)

	attrs.Range(func(
		k string,
		v pcommon.Value,
	) bool {

		result[k] = v.AsString()

		return true
	})

	return result
}

// convertPEvents converts OpenTelemetry span events
// into SarosObserv span events.
func convertPEvents(
	events ptrace.SpanEventSlice,
) []telemetry.SpanEvent {

	result := make(
		[]telemetry.SpanEvent,
		events.Len(),
	)

	for i := 0; i < events.Len(); i++ {
		e := events.At(i)

		result[i] = telemetry.SpanEvent{
			Timestamp: e.Timestamp().AsTime(),
			Name:      e.Name(),

			Attributes: convertPAttributes(
				e.Attributes(),
			),
		}
	}

	return result
}
