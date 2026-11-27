// Use relative URLs to go through nginx proxy
const API_BASE = '/api';

export interface Service {
  name: string;
  span_count: number;
}

export interface TraceSummary {
  trace_id: string;
  service_name: string;
  operation_name: string;
  start_time: string;
  duration_ns: number;
  span_count: number;
}

export interface Span {
  TraceID: string;
  SpanID: string;
  ParentSpanID: string;
  ServiceName: string;
  OperationName: string;
  StartTime: string;
  EndTime: string;
  DurationNs: number;
  Status: string;
  StatusMessage: string;
  Attributes: Record<string, string>;
}

export interface TraceResponse {
  trace_id: string;
  spans: Span[];
}

export interface RequestBucket {
  start: string;
  requests: number;
  errors: number;
}

export interface RequestOverview {
  buckets: RequestBucket[];
  failed: { trace_id: string; span_id: string; service_name: string; operation: string; start: string }[];
}

export interface TraceLog {
  id: string;
  timestamp: string;
  service_name: string;
  span_id: string;
  severity: string;
  message: string;
  attributes: Record<string, string>;
}

export const api = {
  async getRequestOverview(signal: AbortSignal): Promise<RequestOverview> {
    const response = await fetch(`${API_BASE}/overview/requests`, { signal });
    if (!response.ok) throw new Error(`Could not load request activity (${response.status})`);
    return response.json();
  },

  async getTraceLogs(traceId: string, signal: AbortSignal): Promise<{ logs: TraceLog[] }> {
    const response = await fetch(`${API_BASE}/traces/${encodeURIComponent(traceId)}/logs`, { signal });
    if (!response.ok) throw new Error(`Could not load related logs (${response.status})`);
    return response.json();
  },

  async getServices(): Promise<{ services: Service[] }> {
    const response = await fetch(`${API_BASE}/services`);
    if (!response.ok) {
      throw new Error(`Failed to fetch services: ${response.statusText}`);
    }
    return response.json();
  },

  async getTraces(limit: number = 50): Promise<{ traces: TraceSummary[] }> {
    const response = await fetch(`${API_BASE}/traces?limit=${limit}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch traces: ${response.statusText}`);
    }
    return response.json();
  },

  async getTrace(traceId: string): Promise<TraceResponse> {
    const response = await fetch(`${API_BASE}/traces/${traceId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch trace: ${response.statusText}`);
    }
    return response.json();
  },
};
