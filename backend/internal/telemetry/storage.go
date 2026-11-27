package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// SpanData represents a single trace span.
type SpanData struct {
	TraceID       string
	SpanID        string
	ParentSpanID  string
	ServiceName   string
	OperationName string
	StartTime     time.Time
	EndTime       time.Time
	DurationNs    uint64
	Kind          string
	Status        string
	StatusMessage string
	Attributes    map[string]string
	Events        []SpanEvent
}

// SpanEvent represents an event attached to a span.
type SpanEvent struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]string
}

// TraceData represents a complete trace.
type TraceData struct {
	TraceID string
	Spans   []SpanData
}

// ServiceInfo represents a service discovered from stored spans.
type ServiceInfo struct {
	Name      string `json:"name"`
	SpanCount uint64 `json:"span_count"`
}

// TraceSummary represents a trace summary for the trace list.
type TraceSummary struct {
	TraceID       string    `json:"trace_id"`
	ServiceName   string    `json:"service_name"`
	OperationName string    `json:"operation_name"`
	StartTime     time.Time `json:"start_time"`
	DurationNs    uint64    `json:"duration_ns"`
	SpanCount     uint64    `json:"span_count"`
}

// Storage defines the storage interface used by SarosObserv.
type Storage interface {
	StoreTrace(ctx context.Context, trace TraceData) error
	StoreSpan(ctx context.Context, span SpanData) error
	StoreLogs(ctx context.Context, logs []LogRecord) error
	ListTraceLogs(ctx context.Context, traceID string) ([]LogRecord, error)
	RequestOverview(ctx context.Context, now time.Time) (RequestOverview, error)

	ListServices(ctx context.Context) ([]ServiceInfo, error)
	ListTraces(ctx context.Context, limit int) ([]TraceSummary, error)
	GetTrace(ctx context.Context, traceID string) ([]SpanData, error)

	Close() error
}

// ClickHouseStorage implements Storage using ClickHouse.
type ClickHouseStorage struct {
	conn clickhouse.Conn
}

// NewClickHouseStorage creates a ClickHouse storage connection.
func NewClickHouseStorage(
	ctx context.Context,
	dsn string,
) (*ClickHouseStorage, error) {

	// dsn is currently provided through the environment,
	// but clickhouse.Open expects connection options rather
	// than a *sql.DB.
	//
	// The backend runs inside Docker, so ClickHouse is reachable
	// through the Docker service name:
	//
	//     clickhouse:9000
	//
	// The database credentials come from docker-compose.yml.

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{
			"clickhouse:9000",
		},
		Auth: clickhouse.Auth{
			Database: "sarosobserv",
			Username: "sarosobserv",
			Password: "sarosobserv",
		},
	})

	if err != nil {
		return nil, fmt.Errorf(
			"failed to open ClickHouse connection: %w",
			err,
		)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf(
			"failed to ping ClickHouse: %w",
			err,
		)
	}

	return &ClickHouseStorage{
		conn: conn,
	}, nil
}

// StoreTrace stores all spans belonging to a trace.
func (s *ClickHouseStorage) StoreTrace(
	ctx context.Context,
	trace TraceData,
) error {
	for _, span := range trace.Spans {
		if err := s.StoreSpan(ctx, span); err != nil {
			return err
		}
	}

	return nil
}

// StoreSpan stores a single span.
func (s *ClickHouseStorage) StoreSpan(
	ctx context.Context,
	span SpanData,
) error {
	query := `
		INSERT INTO sarosobserv.spans
		(
			trace_id,
			span_id,
			parent_span_id,
			service_name,
			operation_name,
			start_time,
			end_time,
			duration_ns,
			kind,
			status,
			status_message,
			attributes,
			events
		)
		VALUES
		(
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	attributes := make(map[string]string)

	for key, value := range span.Attributes {
		attributes[key] = value
	}

	events := make([]string, 0, len(span.Events))

	for _, event := range span.Events {
		events = append(events, event.Name)
	}

	if err := s.conn.Exec(
		ctx,
		query,
		span.TraceID,
		span.SpanID,
		span.ParentSpanID,
		span.ServiceName,
		span.OperationName,
		span.StartTime,
		span.EndTime,
		span.DurationNs,
		span.Kind,
		span.Status,
		span.StatusMessage,
		attributes,
		events,
	); err != nil {
		return fmt.Errorf(
			"failed to store span: %w",
			err,
		)
	}

	return nil
}

// ListServices returns services discovered from stored spans.
func (s *ClickHouseStorage) ListServices(
	ctx context.Context,
) ([]ServiceInfo, error) {

	query := `
		SELECT
			service_name,
			count() AS span_count
		FROM sarosobserv.spans
		WHERE service_name != ''
		GROUP BY service_name
		ORDER BY span_count DESC
	`

	rows, err := s.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query services: %w",
			err,
		)
	}

	defer rows.Close()

	var services []ServiceInfo

	for rows.Next() {
		var service ServiceInfo

		if err := rows.Scan(
			&service.Name,
			&service.SpanCount,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan service: %w",
				err,
			)
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate services: %w",
			err,
		)
	}

	return services, nil
}

// ListTraces returns recent trace summaries.
func (s *ClickHouseStorage) ListTraces(
	ctx context.Context,
	limit int,
) ([]TraceSummary, error) {

	if limit <= 0 {
		limit = 50
	}

	if limit > 500 {
		limit = 500
	}

	query := `
		SELECT
			trace_id,
			any(service_name) AS service_name,
			any(operation_name) AS operation_name,
			min(start_time) AS start_time,
			max(end_time) AS end_time,
			count() AS span_count
		FROM sarosobserv.spans
		GROUP BY trace_id
		ORDER BY start_time DESC
		LIMIT ?
	`

	rows, err := s.conn.Query(
		ctx,
		query,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to query traces: %w",
			err,
		)
	}

	defer rows.Close()

	var traces []TraceSummary

	for rows.Next() {
		var trace TraceSummary
		var endTime time.Time

		if err := rows.Scan(
			&trace.TraceID,
			&trace.ServiceName,
			&trace.OperationName,
			&trace.StartTime,
			&endTime,
			&trace.SpanCount,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan trace: %w",
				err,
			)
		}

		if endTime.After(trace.StartTime) {
			trace.DurationNs = uint64(
				endTime.Sub(trace.StartTime).Nanoseconds(),
			)
		}

		traces = append(
			traces,
			trace,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate traces: %w",
			err,
		)
	}

	return traces, nil
}

// GetTrace returns all spans belonging to a trace.
func (s *ClickHouseStorage) GetTrace(
	ctx context.Context,
	traceID string,
) ([]SpanData, error) {

	query := `
		SELECT
			trace_id,
			span_id,
			parent_span_id,
			service_name,
			operation_name,
			start_time,
			end_time,
			duration_ns,
			kind,
			status,
			status_message,
			attributes
		FROM sarosobserv.spans
		WHERE trace_id = ?
		ORDER BY start_time ASC
	`

	rows, err := s.conn.Query(
		ctx,
		query,
		traceID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to query trace: %w",
			err,
		)
	}

	defer rows.Close()

	var spans []SpanData

	for rows.Next() {
		var span SpanData

		if err := rows.Scan(
			&span.TraceID,
			&span.SpanID,
			&span.ParentSpanID,
			&span.ServiceName,
			&span.OperationName,
			&span.StartTime,
			&span.EndTime,
			&span.DurationNs,
			&span.Kind,
			&span.Status,
			&span.StatusMessage,
			&span.Attributes,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan span: %w",
				err,
			)
		}

		spans = append(
			spans,
			span,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate trace: %w",
			err,
		)
	}

	return spans, nil
}

// Close closes the ClickHouse connection.
func (s *ClickHouseStorage) Close() error {
	if s.conn == nil {
		return nil
	}

	return s.conn.Close()
}
