import useSWR from "swr";
import { TraceDetail } from "../types/types";
import {
  Badge,
  ScrollArea,
  Stack,
  Text,
  Table,
  Title,
  useMantineTheme,
} from "@mantine/core";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { api } from "../utils/api";
import { useSecret } from "../context/SecretContext";
import { Drawer } from "./Drawer/Drawer";

export interface TraceDetailViewProps {
  traceId: string;
  onClose: () => void;
}

export function TraceDetailView({ traceId, onClose }: TraceDetailViewProps) {
  const { secret } = useSecret();
  const fetcher = async (url: string) => {
    const res = await api.get(url, { params: { secret } });
    return res.data;
  };
  const { data, error } = useSWR<TraceDetail>(
    secret && `/traces/${traceId}`,
    fetcher,
  );
  const theme = useMantineTheme();

  if (error) return <div>Failed to load trace details</div>;
  if (!data) return <div>Loading...</div>;

  const { spans } = data;
  const traceStartTime = new Date(spans[0].start_time).getTime();

  const chartData = spans.map((span) => {
    const spanStartTime = new Date(span.start_time).getTime();
    const spanEndTime = new Date(span.end_time).getTime();
    return {
      name: span.name,
      service: span.service_name,
      duration: (spanEndTime - spanStartTime) / 1000, // in ms
      startTime: (spanStartTime - traceStartTime) / 1000, // in ms
      hasError: span.has_error,
    };
  });

  return (
    <Drawer
      opened
      onClose={onClose}
      title={<Text fw="500">Trace: {traceId}</Text>}
    >
      <Stack gap="xl" p="md">
        <Title order={4} fw={500}>
          Spans Timeline
        </Title>
        <ResponsiveContainer width="100%" height={400} style={{ fontSize: 12 }}>
          <BarChart data={chartData} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" domain={["dataMin", "dataMax"]} />
            <YAxis dataKey="name" type="category" width={150} />
            <Tooltip
              formatter={(value: any, name: any) => {
                if (name === "duration") {
                  return `${value.toFixed(2)}ms`;
                }
                return value;
              }}
            />
            <Legend />
            <Bar dataKey="startTime" stackId="a" fill={theme.colors.gray[5]} />
            <Bar dataKey="duration" stackId="a" fill={theme.colors.teal[5]} />
          </BarChart>
        </ResponsiveContainer>

        <Title order={4} fw={500}>
          Spans Details
        </Title>
        <ScrollArea style={{ height: 400 }}>
          <Table verticalSpacing="sm" fz="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Span</Table.Th>
                <Table.Th>Service</Table.Th>
                <Table.Th>Duration</Table.Th>
                <Table.Th>Status</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {spans.map((span) => (
                <Table.Tr key={span.span_id}>
                  <Table.Td>
                    <Text fw={500} inherit>
                      {span.name}
                    </Text>
                    <Text inherit c="dimmed" fz="xs">
                      {span.span_id}
                    </Text>
                  </Table.Td>
                  <Table.Td>{span.service_name}</Table.Td>
                  <Table.Td>
                    {(
                      new Date(span.end_time).getTime() -
                      new Date(span.start_time).getTime()
                    ).toFixed(2)}
                    ms
                  </Table.Td>
                  <Table.Td>
                    <Badge
                      color={span.has_error ? "red" : "green"}
                      variant="light"
                    >
                      {span.has_error ? "Error" : "OK"}
                    </Badge>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </ScrollArea>
      </Stack>
    </Drawer>
  );
}
