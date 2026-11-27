// Command investigation exports sample checkout requests and correlated logs.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	endpoint := flag.String("endpoint", "127.0.0.1:4317", "local OTLP/gRPC address")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := export(ctx, *endpoint); err != nil {
		log.Fatal(err)
	}
}

func export(ctx context.Context, endpoint string) error {
	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()

	traceRequest := ptraceotlp.NewExportRequest()
	traceResource := traceRequest.Traces().ResourceSpans().AppendEmpty()
	traceResource.Resource().Attributes().PutStr("service.name", "checkout-demo")
	spans := traceResource.ScopeSpans().AppendEmpty().Spans()

	logRequest := plogotlp.NewExportRequest()
	logResource := logRequest.Logs().ResourceLogs().AppendEmpty()
	logResource.Resource().Attributes().PutStr("service.name", "checkout-demo")
	logs := logResource.ScopeLogs().AppendEmpty().LogRecords()

	var failed []pcommon.TraceID
	now := time.Now()
	for minute := 0; minute < 12; minute++ {
		when := now.Add(-time.Duration(12-minute) * time.Minute)
		for request := 0; request < 3+minute%4; request++ {
			var traceID pcommon.TraceID
			var spanID pcommon.SpanID
			if _, err := rand.Read(traceID[:]); err != nil {
				return err
			}
			if _, err := rand.Read(spanID[:]); err != nil {
				return err
			}
			span := spans.AppendEmpty()
			span.SetTraceID(traceID)
			span.SetSpanID(spanID)
			span.SetKind(ptrace.SpanKindServer)
			span.SetName("POST /checkout")
			span.SetStartTimestamp(pcommon.NewTimestampFromTime(when))
			end := pcommon.NewTimestampFromTime(when.Add(180 * time.Millisecond))
			span.SetEndTimestamp(end)
			if minute < 8 || request >= 2 {
				continue
			}
			span.Status().SetCode(ptrace.StatusCodeError)
			span.Status().SetMessage("payment provider timeout")
			entry := logs.AppendEmpty()
			entry.SetTraceID(traceID)
			entry.SetSpanID(spanID)
			entry.SetTimestamp(end)
			entry.SetSeverityNumber(plog.SeverityNumberError)
			entry.Body().SetStr("Payment authorization timed out")
			entry.Attributes().PutStr("payment.provider", "demo")
			failed = append(failed, traceID)
		}
	}

	if _, err := ptraceotlp.NewGRPCClient(conn).Export(ctx, traceRequest); err != nil {
		return fmt.Errorf("export checkout traces: %w", err)
	}
	if _, err := plogotlp.NewGRPCClient(conn).Export(ctx, logRequest); err != nil {
		return fmt.Errorf("export checkout logs: %w", err)
	}

	fmt.Printf("Exported %d requests and %d correlated error logs for checkout-demo\n", spans.Len(), logs.Len())
	for _, traceID := range failed {
		fmt.Println(hex.EncodeToString(traceID[:]))
	}
	return nil
}
