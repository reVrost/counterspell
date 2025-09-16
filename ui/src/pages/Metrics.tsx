import { Stack } from "@mantine/core";
import { useState } from "react";
import { TracesTable } from "../components/TracesTable";
import { TraceDetailView } from "../components/TraceDetail";

export default function MetricsPage() {
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);

  return (
    <>
      <Stack gap="xl">
        <TracesTable
          onTraceClick={(trace) => setSelectedTraceId(trace.trace_id)}
        />
      </Stack>
      {selectedTraceId && (
        <TraceDetailView
          traceId={selectedTraceId}
          onClose={() => setSelectedTraceId(null)}
        />
      )}
    </>
  );
}
