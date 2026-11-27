package ingestion

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sarosobserv/sarosobserv/backend/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logService struct {
	plogotlp.UnimplementedGRPCServer
	storage telemetry.Storage
}

func (s *logService) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	response := plogotlp.NewExportResponse()
	resources := req.Logs().ResourceLogs()
	received := time.Now()
	var records []telemetry.LogRecord
	for i := 0; i < resources.Len(); i++ {
		resource := resources.At(i)
		service := "unknown"
		if name, ok := resource.Resource().Attributes().Get("service.name"); ok && name.Str() != "" {
			service = name.Str()
		}
		scopes := resource.ScopeLogs()
		for j := 0; j < scopes.Len(); j++ {
			logs := scopes.At(j).LogRecords()
			for k := 0; k < logs.Len(); k++ {
				record := logs.At(k)
				timestamp := record.Timestamp()
				if timestamp == 0 {
					timestamp = record.ObservedTimestamp()
				}
				when := received
				if timestamp != 0 {
					when = timestamp.AsTime()
				}
				severity := record.SeverityText()
				if severity == "" && record.SeverityNumber() != 0 {
					severity = record.SeverityNumber().String()
				}
				records = append(records, telemetry.LogRecord{
					ID: uuid.NewString(), TraceID: record.TraceID().String(), SpanID: record.SpanID().String(),
					ServiceName: service, Severity: severity, Message: record.Body().AsString(),
					Attributes: convertPAttributes(record.Attributes()), Timestamp: when,
				})
			}
		}
	}
	if err := s.storage.StoreLogs(ctx, records); err != nil {
		log.Printf("Store OTLP logs: %v", err)
		return response, status.Error(codes.Unavailable, "could not store logs")
	}
	return response, nil
}
