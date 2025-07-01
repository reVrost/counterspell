import { Button, Container, Group, Input, Stack, Title } from "@mantine/core";
import { useLocalStorage } from "@mantine/hooks";
import { notifications } from "@mantine/notifications";
import { IconCheck, IconDeviceFloppy } from "@tabler/icons-react";

export default function SettingsPage() {
  const [secret, setSecret] = useLocalStorage<string>({
    key: "secret-token",
    defaultValue: "",
  });

  return (
    <Container p="xl" size="responsive">
      <Title fw={500} mb="xl">
        Settings
      </Title>
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
  );
}
