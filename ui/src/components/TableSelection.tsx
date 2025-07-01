import { useEffect, useState } from "react";
import cx from "clsx";
import { Checkbox, Group, ScrollArea, Table, Text } from "@mantine/core";
import classes from "./TableSelection.module.css";
import { IconArrowRight, IconX } from "@tabler/icons-react";
import { type ApiResponse, Log } from "../utils/types";
import { api } from "../utils/api";
import useSWR from "swr";
import { notifications } from "@mantine/notifications";
import { useLocalStorage } from "@mantine/hooks";

export function TableSelection() {
  const [secret] = useLocalStorage<string>({
    key: "secret-token",
    defaultValue: "",
  });
  const fetcher = async (url: string) => {
    const res = await api.get(url, { params: { secret } });
    return res.data;
  };

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

  const rows = data?.map((item) => {
    const selected = selection.includes(item.id);
    return (
      <Table.Tr
        key={item.id}
        className={cx(classes.row, { [classes.rowSelected]: selected })}
      >
        <Table.Td>
          <Checkbox
            checked={selection.includes(item.id)}
            onChange={() => toggleRow(item.id)}
          />
        </Table.Td>
        <Table.Td>
          <Group gap="sm">
            <Text inherit fw={500}>
              {item.level}
            </Text>
          </Group>
        </Table.Td>
        <Table.Td>{item.message}</Table.Td>
        <Table.Td>
          <Group justify="space-between">
            {item.timestamp}

            <IconArrowRight />
          </Group>
        </Table.Td>
      </Table.Tr>
    );
  });

  return (
    <ScrollArea>
      <Table miw={800} verticalSpacing="xs" fw={400} variant="compact" fz="xs">
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
            <Table.Th w={200}>Level</Table.Th>
            <Table.Th>Message</Table.Th>
            <Table.Th w={250}>Created</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>{rows}</Table.Tbody>
      </Table>
    </ScrollArea>
  );
}
