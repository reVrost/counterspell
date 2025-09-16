import { useEffect, useState } from "react";
import {
  ActionIcon,
  Code,
  Group,
  ScrollArea,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
} from "@mantine/core";
import { IconRefresh, IconSearch, IconX } from "@tabler/icons-react";
import { type ApiResponse, Log } from "../utils/types";
import { api } from "../utils/api";
import useSWR, { mutate } from "swr";
import { notifications } from "@mantine/notifications";
import { useDisclosure } from "@mantine/hooks";
import { useSearchParams } from "react-router-dom";
import { useSecret } from "../context/SecretContext";
import { OpsTable } from "./OpsTable/OpsTable";
import { Drawer } from "./Drawer/Drawer";

export function LogsTable() {
  const { secret } = useSecret();
  const [searchParams, setSearchParams] = useSearchParams();
  const [filter, setFilter] = useState(searchParams.get("q") || "");
  const [level, setLevel] = useState(searchParams.get("level") || "");

  const fetcher = async (url: string) => {
    const res = await api.get(url, {
      params: { secret, q: filter, level },
    });
    return res.data;
  };

  const { data: response, error } = useSWR<ApiResponse<Log>>(
    [`/logs?q=${filter}&level=${level}`, secret],
    async ([url, secret]) => {
      if (!secret) {
        // SWR will catch this error and return it in the `error` object.
        throw new Error("Secret is not set.");
      }
      return fetcher(url);
    },
  );
  const data = response?.data || [];

  // const { data: response, error } = useSWR<ApiResponse<Log>>(
  //   secret ? `/logs?q=${filter}&level=${level}` : null,
  //   fetcher,
  // );
  // const data = response?.data || [];

  useEffect(() => {
    if (error) {
      notifications.show({
        title: "Error fetching logs",
        message:
          "There was an error fetching the logs. Reason: " + error.message,
        color: "red",
        icon: <IconX />,
      });
    }
  }, [error]);

  useEffect(() => {
    setSearchParams({ q: filter, level });
  }, [filter, level, setSearchParams]);

  const [opened, { open, close }] = useDisclosure(false);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);

  const handleRowClick = (log: Log) => {
    setSelectedLog(log);
    open();
  };

  const levelToColor: Record<string, string> = {
    debug: "var(--mantine-color-teal-5)",
    info: "var(--mantine-color-blue-5)",
    warn: "var(--mantine-color-orange-5)",
    error: "var(--mantine-color-red-5)",
  };

  const rows = data?.map((item) => {
    return (
      <Table.Tr
        key={item.id}
        style={{
          cursor: "pointer",
          "&:hover": {
            backgroundColor: "var(--mantine-color-gray-2)",
          },
        }}
        onClick={() => handleRowClick(item)}
      >
        <Table.Td>{new Date(item.timestamp).toLocaleString()}</Table.Td>
        <Table.Td tt="uppercase" c={levelToColor[item.level]}>
          {item.level}
        </Table.Td>
        <Table.Td>{item.message}</Table.Td>
        <Table.Td>{JSON.stringify(item.attributes)}</Table.Td>
      </Table.Tr>
    );
  });

  const flattenAttributes = (
    attributes: Record<string, any>,
    prefix = "attribute.",
  ) => {
    let flat: Record<string, any> = {};
    for (const key in attributes) {
      if (typeof attributes[key] === "object" && attributes[key] !== null) {
        Object.assign(
          flat,
          flattenAttributes(attributes[key], `${prefix}${key}.`),
        );
      } else {
        flat[`${prefix}${key}`] = attributes[key];
      }
    }
    return flat;
  };

  const drawerData = selectedLog
    ? {
        id: selectedLog.id,
        message: selectedLog.message,
        timestamp: selectedLog.timestamp,
        ...flattenAttributes(selectedLog.attributes),
      }
    : {};

  const handleSearch = () => {
    mutate(`/logs?q=${filter}&level=${level}`);
  };

  return (
    <ScrollArea>
      <Stack>
        <Group p="xs">
          <TextInput
            rightSection={<IconSearch size={18} stroke={1.5} />}
            placeholder="Search term or filter"
            value={filter}
            onChange={(event) => setFilter(event.currentTarget.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                handleSearch();
              }
            }}
          />

          <Select
            placeholder="Log level"
            value={level}
            onChange={(value) => setLevel(value || "")}
            data={[
              { value: "debug", label: "Debug" },
              { value: "info", label: "Info" },
              { value: "warn", label: "Warn" },
              { value: "error", label: "Error" },
            ]}
          />
          <ActionIcon variant="light" onClick={() => mutate("logs")}>
            <IconRefresh />
          </ActionIcon>
        </Group>
        <OpsTable
          columns={["Time", "Level", "Message", "Attributes"]}
          rows={rows}
        />
      </Stack>
      <Drawer opened={opened} onClose={close} title="Log Details">
        {selectedLog && (
          <Table withTableBorder fz="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Property</Table.Th>
                <Table.Th>Value</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {Object.entries(drawerData).map(([key, value]) => (
                <Table.Tr key={key}>
                  <Table.Td>{key}</Table.Td>
                  <Table.Td>
                    <Text inherit>{value as string}</Text>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Drawer>
    </ScrollArea>
  );
}
