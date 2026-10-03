package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sarosobserv/sarosobserv/backend/internal/api"
	"github.com/sarosobserv/sarosobserv/backend/internal/ingestion"
	"github.com/sarosobserv/sarosobserv/backend/internal/telemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize telemetry storage
	storage, err := telemetry.NewClickHouseStorage(ctx, "clickhouse://sarosobserv:sarosobserv@localhost:9000/sarosobserv")
	if err != nil {
		log.Fatalf("Failed to initialize ClickHouse storage: %v", err)
	}
	defer storage.Close()

	// Initialize OTLP ingestion server
	otlpServer := ingestion.NewOTLPServer(storage, ":4317", ":4318")

	// Start OTLP servers
	go func() {
		log.Println("Starting OTLP gRPC server on :4317")
		if err := otlpServer.StartGRPC(ctx); err != nil {
			log.Printf("OTLP gRPC server error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting OTLP HTTP server on :4318")
		if err := otlpServer.StartHTTP(ctx); err != nil {
			log.Printf("OTLP HTTP server error: %v", err)
		}
	}()

	// Initialize REST API
	router := gin.Default()
	apiHandler := api.NewHandler(storage)
	apiHandler.RegisterRoutes(router)

	// Start REST API
	apiAddr := ":8080"
	apiListener, err := net.Listen("tcp", apiAddr)
	if err != nil {
		log.Fatalf("Failed to create API listener: %v", err)
	}

	go func() {
		log.Printf("Starting SarosObserv API on %s", apiAddr)
		if err := router.RunListener(apiListener); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down SarosObserv...")
	cancel()
	log.Println("SarosObserv stopped")
}
