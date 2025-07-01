import { useEffect, useState } from "react";
import cx from "clsx";
import {
  Badge,
  Checkbox,
  Code,
  Drawer,
  Text,
  Group,
  ScrollArea,
  Table,
  useMantineTheme,
} from "@mantine/core";
import classes from "./LogTable.module.css";
import { IconChevronRight, IconX } from "@tabler/icons-react";
import { type ApiResponse, Log } from "../utils/types";
import { api } from "../utils/api";
import useSWR from "swr";
import { notifications } from "@mantine/notifications";
import { useDisclosure, useLocalStorage } from "@mantine/hooks";

export function LogTable() {
  const [secret] = useLocalStorage<string>({
    key: "secret-token",
    defaultValue: "",
  });
  const fetcher = async (url: string) => {
    const res = await api.get(url, { params: { secret } });
    return res.data;
  };
  const theme = useMantineTheme();

  const { data: response, error } = useSWR<ApiResponse<Log>>(
    secret && "/logs",
    fetcher,
  );
  const data = response?.data || [];

  useEffect(() => {
    if (error) {
      notifications.show({
        title: "Error fetching logs",
        message:
          "There was an error fetching the logs. Please try again later.",
        color: "red",
        icon: <IconX />,
      });
    }
  }, [error]);

  const [selection, setSelection] = useState<string[]>([]);
  const [opened, { open, close }] = useDisclosure(false);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);

  const handleRowClick = (log: Log) => {
    setSelectedLog(log);
    open();
  };

  const toggleRow = (id: string) =>
    setSelection((current) =>
      current.includes(id)
        ? current.filter((item) => item !== id)
        : [...current, id],
    );
  const toggleAll = () =>
    setSelection((current) =>
      current.length === data?.length ? [] : data?.map((item) => item.id) || [],
    );

  const levelToIcon: Record<string, string> = {
    debug: "var(--mantine-color-teal-4)",
    info: "var(--mantine-color-blue-4)",
    warn: "var(--mantine-color-orange-4)",
    error: "var(--mantine-color-red-4)",
  };

  const rows = data?.map((item) => {
    const selected = selection.includes(item.id);
    return (
      <Table.Tr
        key={item.id}
        className={cx(classes.row, { [classes.rowSelected]: selected })}
        onClick={() => handleRowClick(item)}
      >
        <Table.Td>
          <Checkbox
            checked={selection.includes(item.id)}
            onChange={() => toggleRow(item.id)}
            onClick={(e) => e.stopPropagation()}
          />
        </Table.Td>
        <Table.Td>
          <Group gap="sm">
            <Badge variant="dot" color={levelToIcon[item.level]}>
              {item.level}
            </Badge>
          </Group>
        </Table.Td>
        <Table.Td>{item.message}</Table.Td>
        <Table.Td>
          <Code color="teal.1">{JSON.stringify(item.attributes)}</Code>
        </Table.Td>
        <Table.Td>
          <Group justify="space-between">
            {item.timestamp}
            <IconChevronRight size={16} color={theme.colors.gray[5]} />
          </Group>
        </Table.Td>
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

  return (
    <>
      <Drawer
        opened={opened}
        onClose={close}
        title="Log Details"
        position="right"
        size="lg"
      >
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
      <ScrollArea>
        <Table
          miw={800}
          verticalSpacing="xs"
          fw={500}
          variant="compact"
          fz="xs"
        >
          <Table.Thead>
            <Table.Tr>
              <Table.Th w={10}>
                <Checkbox
                  onChange={toggleAll}
                  checked={selection.length === data?.length}
                  indeterminate={
                    selection.length > 0 && selection.length !== data?.length
                  }
                />
              </Table.Th>
              <Table.Th w={150}>Level</Table.Th>
              <Table.Th>Message</Table.Th>
              <Table.Th>Attributes</Table.Th>
              <Table.Th w={250}>Created</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>{rows}</Table.Tbody>
        </Table>
      </ScrollArea>
    </>
  );
}
