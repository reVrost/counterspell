import {
  Button,
  Group,
  Stack,
  Text,
  Title,
  Grid,
  Paper,
  Select,
} from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";
import { Link } from "react-router-dom";
import Chat from "../components/Chat/Chat";

export default function AgentsPage() {
  return (
    <Grid h="100%">
      <Grid.Col span={12}>
        <Stack h="100%" p="lg">
          <Group justify="space-between">
            <Stack gap="4xs">
              <Title>Agents</Title>
              <Text size="sm" c="dimmed">
                Select a blueprint and start a conversation.
              </Text>
            </Stack>
            <Button
              component={Link}
              to="/agents/create"
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
          <Select
            label="Select a blueprint"
            placeholder="Pick a blueprint to start"
            data={[]}
            mt="md"
          />
          <Chat />
        </Stack>
      </Grid.Col>
    </Grid>
  );
}
