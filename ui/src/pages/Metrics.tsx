import { ActionIcon, Container, Group, Stack, Title } from "@mantine/core";
import { IconRefresh } from "@tabler/icons-react";
import { useState } from "react";
import { TracesTable } from "../components/TracesTable";
import { TraceDetailView } from "../components/TraceDetail";

export default function MetricsPage() {
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);

  return (
    <Container p="xl" size="responsive">
      <Group justify="space-between" align="center">
        <Title fw="bold" mb="xl">
          Metrics
        </Title>
        <ActionIcon variant="subtle">
          <IconRefresh />
        </ActionIcon>
      </Group>
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
  );
}
