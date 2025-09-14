import { Button, Container, Group, Input, Stack, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconCheck, IconDeviceFloppy } from "@tabler/icons-react";
import { useSecret } from "../context/SecretContext";

export default function SettingsPage() {
  const { secret, setSecret } = useSecret();

  return (
    <>
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
