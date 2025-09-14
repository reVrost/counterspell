import { Group, Title } from "@mantine/core";

export const MainHeader = () => (
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
    <Title order={3}>Header</Title>
  </Group>
);
