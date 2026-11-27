-- SarosObserv ClickHouse Schema
-- This file initializes the database schema for SarosObserv

-- Create database
CREATE DATABASE IF NOT EXISTS sarosobserv;

USE sarosobserv;

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
    trace_id String,
    service_name String,
    start_time DateTime64(9),
    end_time DateTime64(9),
    duration_ns UInt64,
    span_count UInt32,
    root_span_name String,
    attributes Map(String, String),
    timestamp DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (trace_id, start_time)
TTL timestamp + INTERVAL 7 DAY;

-- Spans table
CREATE TABLE IF NOT EXISTS spans (
    span_id String,
    trace_id String,
    parent_span_id String,
    service_name String,
    operation_name String,
    start_time DateTime64(9),
    end_time DateTime64(9),
    duration_ns UInt64,
    kind String,
    status String,
    status_message String,
    attributes Map(String, String),
    events Array(Tuple(DateTime64(9), String, Map(String, String))),
    timestamp DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (trace_id, span_id, start_time)
TTL timestamp + INTERVAL 7 DAY;

-- Metrics table (for Phase 4)
CREATE TABLE IF NOT EXISTS metrics (
    metric_name String,
    service_name String,
    metric_type String,
    value Float64,
    attributes Map(String, String),
    timestamp DateTime64(9)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (metric_name, service_name, timestamp)
TTL toDateTime(timestamp) + INTERVAL 30 DAY;

-- Logs table (for Phase 5)
CREATE TABLE IF NOT EXISTS logs (
    log_id String,
    trace_id String,
    span_id String,
    service_name String,
    severity String,
    message String,
    attributes Map(String, String),
    timestamp DateTime64(9)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, service_name)
TTL toDateTime(timestamp) + INTERVAL 14 DAY;

-- Service discovery table
CREATE TABLE IF NOT EXISTS services (
    service_name String,
    first_seen DateTime,
    last_seen DateTime,
    attributes Map(String, String),
    timestamp DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree()
ORDER BY service_name;

-- Endpoints table
CREATE TABLE IF NOT EXISTS endpoints (
    service_name String,
    endpoint_name String,
    first_seen DateTime,
    last_seen DateTime,
    attributes Map(String, String),
    timestamp DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree()
ORDER BY (service_name, endpoint_name);

-- Dependencies table
CREATE TABLE IF NOT EXISTS dependencies (
    source_service String,
    target_service String,
    dependency_type String,
    first_seen DateTime,
    last_seen DateTime,
    attributes Map(String, String),
    timestamp DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree()
ORDER BY (source_service, target_service);
