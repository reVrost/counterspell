import { Button, Group, Stack, Text, Title } from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";

export default function AgentsPage() {
  return (
    <Stack p="lg">
      <Group justify="space-between">
        <Stack gap="4xs">
          <Title>Blueprints</Title>
          <Text size="sm" c="dimmed">
            Blueprints are a collection of agents that can be used to create
            counterspells.
          </Text>
        </Stack>
        <Button
          leftSection={
            <IconPlus
              size={16}
              stroke={1.5}
              style={{
                marginInlineEnd: "-0.45rem",
              }}
            />
          }
        >
          Create Blueprint
        </Button>
      </Group>
    </Stack>
  );
}
