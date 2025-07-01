import {
  ActionIcon,
  Container,
  Group,
  Stack,
  Title,
} from "@mantine/core";
import { IconRefresh } from "@tabler/icons-react";
import { LogTable } from "../components/LogTable";

export default function LogsPage() {
  return (
    <Container p="xl" size="responsive">
      <Group justify="space-between" align="center">
        <Title fw={500} mb="xl">
          Logs
        </Title>
        <ActionIcon variant="subtle">
          <IconRefresh />
        </ActionIcon>
      </Group>
      <Stack gap="xl">
        <LogTable />
      </Stack>
    </Container>
  );
}
