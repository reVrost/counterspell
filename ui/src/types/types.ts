
export interface APIResponse<T> {
  metadata: {
    total: number;
    limit: number;
    offset: number;
  };
  data: T;
}

export interface TraceListItem {
  trace_id: string;
  root_span_name: string;
  trace_start_time: string;
  duration_ms: number;
  span_count: number;
  error_count: number;
  has_error: boolean;
}

export interface SpanItem {
  span_id: string;
  trace_id: string;
  parent_span_id?: string;
  name: string;
  start_time: string;
  end_time: string;
  duration_ns: number;
  attributes: Record<string, any>;
  service_name: string;
  has_error: boolean;
}

export interface TraceDetail {
  trace_id: string;
  spans: SpanItem[];
}
