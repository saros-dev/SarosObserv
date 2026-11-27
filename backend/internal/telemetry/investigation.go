package telemetry

import (
	"context"
	"time"
)

type LogRecord struct {
	ID          string            `json:"id"`
	TraceID     string            `json:"trace_id"`
	SpanID      string            `json:"span_id"`
	ServiceName string            `json:"service_name"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	Attributes  map[string]string `json:"attributes"`
	Timestamp   time.Time         `json:"timestamp"`
}

func (s *ClickHouseStorage) StoreLogs(ctx context.Context, logs []LogRecord) error {
	if len(logs) == 0 {
		return nil
	}
	batch, err := s.conn.PrepareBatch(ctx, `
		INSERT INTO sarosobserv.logs
		(log_id, trace_id, span_id, service_name, severity, message, attributes, timestamp)
	`)
	if err != nil {
		return err
	}
	defer batch.Abort()
	for _, record := range logs {
		if err := batch.Append(record.ID, record.TraceID, record.SpanID, record.ServiceName,
			record.Severity, record.Message, record.Attributes, record.Timestamp); err != nil {
			return err
		}
	}
	return batch.Send()
}

func (s *ClickHouseStorage) ListTraceLogs(ctx context.Context, traceID string) ([]LogRecord, error) {
	rows, err := s.conn.Query(ctx, `
		SELECT log_id, trace_id, span_id, service_name, severity, message, attributes, timestamp
		FROM sarosobserv.logs
		WHERE trace_id = ? AND timestamp >= now() - INTERVAL 14 DAY
		ORDER BY timestamp DESC, log_id
		LIMIT 100
	`, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]LogRecord, 0)
	for rows.Next() {
		var record LogRecord
		if err := rows.Scan(&record.ID, &record.TraceID, &record.SpanID, &record.ServiceName,
			&record.Severity, &record.Message, &record.Attributes, &record.Timestamp); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

type RequestBucket struct {
	Start    time.Time `json:"start"`
	Requests uint64    `json:"requests"`
	Errors   uint64    `json:"errors"`
}

type FailedRequest struct {
	TraceID     string    `json:"trace_id"`
	SpanID      string    `json:"span_id"`
	ServiceName string    `json:"service_name"`
	Operation   string    `json:"operation"`
	Start       time.Time `json:"start"`
}

type RequestOverview struct {
	Buckets []RequestBucket `json:"buckets"`
	Failed  []FailedRequest `json:"failed"`
}

func (s *ClickHouseStorage) RequestOverview(ctx context.Context, now time.Time) (RequestOverview, error) {
	const bucketCount = 60
	start := now.UTC().Truncate(time.Minute).Add(-(bucketCount - 1) * time.Minute)
	result := RequestOverview{
		Buckets: make([]RequestBucket, bucketCount),
		Failed:  make([]FailedRequest, 0),
	}
	for i := range result.Buckets {
		result.Buckets[i].Start = start.Add(time.Duration(i) * time.Minute)
	}

	// Count each server span once, including when the exporter retries it.
	const requests = `
		SELECT trace_id, span_id, min(start_time) AS request_start,
			any(service_name) AS service, any(operation_name) AS operation,
			max(lower(status) = 'error') AS failed
		FROM sarosobserv.spans
		WHERE kind = 'Server' AND start_time >= ? AND start_time <= ?
		GROUP BY trace_id, span_id
	`
	if err := s.loadRequestBuckets(ctx, requests, start, now, result.Buckets); err != nil {
		return RequestOverview{}, err
	}
	rows, err := s.conn.Query(ctx, `
		SELECT trace_id, span_id, service, operation, request_start FROM (`+requests+`)
		WHERE failed = 1
		ORDER BY request_start DESC, trace_id, span_id LIMIT 25
	`, start, now)
	if err != nil {
		return RequestOverview{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var request FailedRequest
		if err := rows.Scan(&request.TraceID, &request.SpanID, &request.ServiceName, &request.Operation, &request.Start); err != nil {
			return RequestOverview{}, err
		}
		result.Failed = append(result.Failed, request)
	}
	return result, rows.Err()
}

func (s *ClickHouseStorage) loadRequestBuckets(ctx context.Context, requests string, start, now time.Time, buckets []RequestBucket) error {
	rows, err := s.conn.Query(ctx, `
		SELECT toStartOfMinute(request_start), count(), sum(failed) FROM (`+requests+`) GROUP BY 1
	`, start, now)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var bucket RequestBucket
		if err := rows.Scan(&bucket.Start, &bucket.Requests, &bucket.Errors); err != nil {
			return err
		}
		index := int(bucket.Start.Sub(start) / time.Minute)
		if index >= 0 && index < len(buckets) {
			buckets[index].Requests = bucket.Requests
			buckets[index].Errors = bucket.Errors
		}
	}
	return rows.Err()
}
