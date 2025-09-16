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
      <Drawer
        opened={opened}
        onClose={close}
        title={
          <Group gap="3xs">
            <Text component="span" c="var(--mantine-color-indigo-4)" inherit>
              Log Details
            </Text>
            <span>
              at{" "}
              {new Date(selectedLog?.timestamp || "").toLocaleString("en-US", {
                month: "short",
                day: "numeric",
                timeZoneName: "short",
                year: "numeric",
                hour: "2-digit",
                minute: "2-digit",
                second: "2-digit",
              })}
            </span>
          </Group>
        }
      >
        {selectedLog && (
          <div
            style={{
              fontFamily: "JetBrains Mono, monospace",
              fontSize: "13px",
              color: "var(--mantine-color-indigo-6)",
              borderTop: "1px solid var(--mantine-color-gray-3)",
            }}
          >
            {JSON.stringify(selectedLog, null, 2)
              .split("\n")
              .map((line, index) => (
                <div style={{ display: "flex" }} key={index}>
                  <span
                    style={{
                      minWidth: "3em",
                      paddingRight: "1em",
                      borderRight: "1px solid var(--mantine-color-gray-3)",
                      textAlign: "right",
                      userSelect: "none",
                      color: "var(--mantine-color-gray-4)",
                    }}
                  >
                    {index + 1}
                  </span>
                  <span style={{ marginLeft: "1em", whiteSpace: "pre" }}>
                    {line}
                  </span>
                </div>
              ))}
          </div>
        )}
      </Drawer>
    </ScrollArea>
  );
}
