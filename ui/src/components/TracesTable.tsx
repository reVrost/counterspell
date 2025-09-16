import { Badge, Stack, Table, Text, TextInput } from "@mantine/core";
import { OpsTable } from "./OpsTable/OpsTable";
import { IconSearch, IconX } from "@tabler/icons-react";
import useSWR, { useSWRConfig } from "swr";
import { api } from "../utils/api";
import { TraceListItem } from "../types/types";
import { useEffect, useState } from "react";
import { useSecret } from "../context/SecretContext";
import { notifications } from "@mantine/notifications";

interface TracesTableProps {
  onTraceClick: (trace: TraceListItem) => void;
}

export function TracesTable({ onTraceClick }: TracesTableProps) {
  const { secret } = useSecret();
  const [filter, setFilter] = useState("");
  const { mutate } = useSWRConfig();

  const fetcher = async (url: string) => {
    const res = await api.get(url, { params: { secret, q: filter } });
    return res.data;
  };

  // TODO: define any type
  const { data, error } = useSWR<any>(
    [`/traces?q=${filter}`, secret],
    async ([url, secret]) => {
      if (!secret) {
        // SWR will catch this error and return it in the `error` object.
        throw new Error("Secret is not set.");
      }
      return fetcher(url);
    },
  );

  useEffect(() => {
    if (error) {
      notifications.show({
        title: "Error fetching metrics",
        message:
          "There was an error fetching metrics data. Reason: " + error.message,
        color: "red",
        icon: <IconX />,
      });
    }
  }, [error]);

  const rows = data?.data.map((trace: TraceListItem) => (
    <Table.Tr
      key={trace.trace_id}
      onClick={() => onTraceClick(trace)}
      style={{ cursor: "pointer" }}
    >
      <Table.Td>{new Date(trace.trace_start_time).toLocaleString()}</Table.Td>
      <Table.Td>
        <Text inherit fw={500}>
          {trace.root_span_name}
        </Text>
        <Text c="dimmed" fz="xs">
          {trace.trace_id}
        </Text>
      </Table.Td>
      <Table.Td>
        <Badge
          variant="dot"
          color={
            trace.has_error
              ? "var(--mantine-color-red-5)"
              : "var(--mantine-color-green-5)"
          }
        >
          {trace.has_error ? "Error" : "OK"}
        </Badge>
      </Table.Td>
      <Table.Td>{trace.duration_ms.toFixed(2)}ms</Table.Td>
      <Table.Td>{trace.span_count}</Table.Td>
    </Table.Tr>
  ));

  const handleSearch = () => {
    mutate(`/traces?q=${filter}`);
  };

  return (
    <Stack gap={0}>
      <TextInput
        radius="md"
        variant="subtle"
        rightSection={<IconSearch size={18} stroke={1.5} />}
        placeholder="Search for a trace by root span name"
        value={filter}
        onChange={(event) => setFilter(event.currentTarget.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            handleSearch();
          }
        }}
      />
      <OpsTable
        columns={["Start Time", "Trace", "Status", "Duration", "Spans"]}
        rows={rows}
      />
    </Stack>
  );
}
