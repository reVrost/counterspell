import { ActionIcon, Container, Group, Stack, Title } from "@mantine/core";
import { useState } from "react";
import { TracesTable } from "../components/TracesTable";
import { TraceDetailView } from "../components/TraceDetail";

export default function MetricsPage() {
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);

  return (
    <>
      <Container p="xl" size="responsive">
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
      </Container>
    </>
  );
}
