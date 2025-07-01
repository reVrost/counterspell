import { Container, Group, Stack, Title } from "@mantine/core";
import { LogTable } from "../components/LogTable";

export default function LogsPage() {
  return (
    <Container p="xl" size="responsive">
      <Group justify="space-between" align="center">
        <Title fw={500} mb="xl">
          Logs
        </Title>
      </Group>
      <Stack gap="xl">
        <LogTable />
      </Stack>
    </Container>
  );
}
