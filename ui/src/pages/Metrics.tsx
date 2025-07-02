import { ActionIcon, Container, Group, Stack, Title } from "@mantine/core";
import { IconRefresh } from "@tabler/icons-react";
import { useState } from "react";
import { TracesTable } from "../components/TracesTable";
import { TraceDetailView } from "../components/TraceDetail";

export default function MetricsPage() {
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);

  return (
    <>
      <Group
        justify="space-between"
        align="center"
        style={{
          backgroundColor: "var(--mantine-color-gray-0)",
          borderBottom: "1px solid var(--mantine-color-gray-2)",
          boxShadow: "0.5px 0.5px  var(--mantine-color-gray-2)",
        }}
        p="12"
      >
        <Title order={2} fw={600}>
          Metrics
        </Title>
        <ActionIcon variant="subtle">
          <IconRefresh />
        </ActionIcon>
      </Group>
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
