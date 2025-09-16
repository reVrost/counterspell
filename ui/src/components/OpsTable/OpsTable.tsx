import { Table, ScrollArea } from "@mantine/core";
import classes from "./OpsTable.module.css";
import React from "react";

interface OpsTableProps<T> {
  columns: string[];
  rows: React.ReactNode;
}

export function OpsTable<T>({ columns, rows }: OpsTableProps<T>) {
  return (
    <Table className={classes.table} variant="compact" fz="xs">
      <Table.Thead
        style={{ borderTop: "1px solid var(--mantine-color-gray-2)" }}
      >
        <Table.Tr>
          {columns.map((col) => (
            <Table.Th key={col}>{col}</Table.Th>
          ))}
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody ff="JetBrains Mono" fw={500}>
        {rows}
      </Table.Tbody>
    </Table>
  );
}
