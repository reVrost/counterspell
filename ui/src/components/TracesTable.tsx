import {
  ActionIcon,
  Badge,
  Table,
  Text,
  TextInput,
  useMantineTheme,
} from "@mantine/core";
import { IconArrowRight, IconChevronRight, IconSearch } from "@tabler/icons-react";
import useSWR, { useSWRConfig } from "swr";
import { api } from "../utils/api";
import { TraceListItem } from "../types/types";
import { useLocalStorage } from "@mantine/hooks";
import { useState } from "react";

interface TracesTableProps {
  onTraceClick: (trace: TraceListItem) => void;
}

export function TracesTable({ onTraceClick }: TracesTableProps) {
  const [secret] = useLocalStorage<string>({
    key: "secret-token",
    defaultValue: "",
  });
  const [filter, setFilter] = useState("");
  const { mutate } = useSWRConfig();

  const fetcher = async (url: string) => {
    const res = await api.get(url, { params: { secret, q: filter } });
    return res.data;
  };
  const { data, error } = useSWR<any>(
    secret ? `/traces?q=${filter}` : null,
    fetcher,
  );
  const theme = useMantineTheme();

  if (error) return <div>Failed to load</div>;
  if (!data) return <div>Loading...</div>;

  const rows = data.data.map((trace: TraceListItem) => (
    <Table.Tr
      key={trace.trace_id}
      onClick={() => onTraceClick(trace)}
      style={{ cursor: "pointer" }}
    >
      <Table.Td>
        <Text inherit fw={500}>
          {trace.root_span_name}
        </Text>
        <Text c="dimmed" fz="xs">
          {trace.trace_id}
        </Text>
      </Table.Td>
      <Table.Td>
        <Badge variant="dot" color={trace.has_error ? "red" : "green"}>
          {trace.has_error ? "Error" : "OK"}
        </Badge>
      </Table.Td>
      <Table.Td>{new Date(trace.trace_start_time).toLocaleString()}</Table.Td>
      <Table.Td>{trace.duration_ms.toFixed(2)}ms</Table.Td>
      <Table.Td>{trace.span_count}</Table.Td>
      <Table.Td>
        <IconChevronRight size={16} color={theme.colors.gray[5]} />
      </Table.Td>
    </Table.Tr>
  ));

  const handleSearch = () => {
    mutate(`/traces?q=${filter}`);
  };

  return (
    <>
      <TextInput
        radius="xl"
        rightSectionWidth={42}
        leftSection={<IconSearch size={18} stroke={1.5} />}
        rightSection={
          <ActionIcon
            size={32}
            radius="xl"
            color={theme.primaryColor}
            variant="filled"
            onClick={handleSearch}
          >
            <IconArrowRight size={18} stroke={1.5} />
          </ActionIcon>
        }
        placeholder="Search for a trace by root span name"
        value={filter}
        onChange={(event) => setFilter(event.currentTarget.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            handleSearch();
          }
        }}
      />
      <Table
        withTableBorder
        verticalSpacing="sm"
        highlightOnHover
        fz="xs"
        fw={500}
      >
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Trace</Table.Th>
            <Table.Th>Status</Table.Th>
            <Table.Th>Start Time</Table.Th>
            <Table.Th>Duration</Table.Th>
            <Table.Th>Spans</Table.Th>
            <Table.Th />
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>{rows}</Table.Tbody>
      </Table>
    </>
  );
}

