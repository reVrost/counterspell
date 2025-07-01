import {
  ActionIcon,
  Container,
  Group,
  Input,
  Stack,
  Title,
  useMantineTheme,
} from "@mantine/core";
import { IconArrowRight, IconRefresh, IconSearch } from "@tabler/icons-react";
import { LogTable } from "../components/LogTable";

export default function LogsPage() {
  const theme = useMantineTheme();
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
        <Input
          radius="xl"
          rightSectionWidth={42}
          leftSection={<IconSearch size={18} stroke={1.5} />}
          rightSection={
            <ActionIcon
              size={32}
              radius="xl"
              color={theme.primaryColor}
              variant="filled"
            >
              <IconArrowRight size={18} stroke={1.5} />
            </ActionIcon>
          }
          placeholder="Search term or filter like 'level > 0'"
        />
        <LogTable />
      </Stack>
    </Container>
  );
}
