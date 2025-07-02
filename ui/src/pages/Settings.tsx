import { Button, Container, Group, Input, Stack, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconCheck, IconDeviceFloppy } from "@tabler/icons-react";
import { useSecret } from "../context/SecretContext";

export default function SettingsPage() {
  const { secret, setSecret } = useSecret();

  return (
    <>
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
        <Title order={2} fw={600}>
          Settings
        </Title>
      </Group>
      <Container p="xl" size="responsive">
        <Stack gap="xl">
          <Input.Wrapper label="Secret Token">
            <Input
              value={secret}
              onChange={(event) => setSecret(event.currentTarget.value)}
              placeholder="Enter your secret token"
            />
          </Input.Wrapper>
          <Group justify="flex-end">
            <Button
              leftSection={<IconDeviceFloppy size={18} />}
              onClick={() => {
                notifications.show({
                  title: "Secret token updated",
                  message: "Your secret token has been updated.",
                  color: "green",
                  icon: <IconCheck />,
                });
              }}
            >
              Save
            </Button>
          </Group>
        </Stack>
      </Container>
    </>
  );
}
