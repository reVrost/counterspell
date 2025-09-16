import { Button, Group, Stack, Text, Title } from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";

export default function AgentsPage() {
  return (
    <Stack p="lg">
      <Group justify="space-between">
        <Stack gap="4xs">
          <Title>Blueprints</Title>
          <Text fz="sm" c="dimmed">
            Create your agentic blueprints here.
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
