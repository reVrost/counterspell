import {
  ActionIcon,
  Container,
  Group,
  Input,
  Stack,
  Title,
  useMantineTheme,
} from "@mantine/core";
import { IconArrowRight, IconRefresh, IconSearch } from "@tabler/icons-react";
import { useState } from "react";
import { TracesTable } from "../components/TracesTable";
import { TraceDetailView } from "../components/TraceDetail";

export default function MetricsPage() {
  const theme = useMantineTheme();
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);

  return (
    <Container p="xl" size="responsive">
      <Group justify="space-between" align="center">
        <Title fw={500} mb="xl">
          Metrics
        </Title>
        <ActionIcon variant="subtle">
          <IconRefresh />
        </ActionIcon>
      </Group>
      <Stack gap="xl">
        <Input
          radius="xl"
          rightSectionWidth={42}
          leftSection={<IconSearch size={18} stroke={1.5} />}
          rightSection={
            <ActionIcon
              size={32}
              radius="xl"
              color={theme.primaryColor}
              variant="filled"
            >
              <IconArrowRight size={18} stroke={1.5} />
            </ActionIcon>
          }
          placeholder="Search for a trace by root span name"
        />
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
