import {
  Avatar,
  Button,
  Group,
  Menu,
  Stack,
  Text,
  useMantineColorScheme,
} from "@mantine/core";
import {
  IconChevronRight,
  IconLogout,
  IconMoon,
  IconSettings,
  IconSun,
} from "@tabler/icons-react";
import { useState } from "react";

export const defaultAvatar =
  "https://raw.githubusercontent.com/mantinedev/mantine/master/.demo/avatars/avatar-8.png";

export function UserButton() {
  const [userMenuOpened, setUserMenuOpened] = useState(false);
  const { colorScheme, setColorScheme } = useMantineColorScheme();

  return (
    <Menu
      width={250}
      position="right"
      transitionProps={{ transition: "pop" }}
      onClose={() => setUserMenuOpened(false)}
      onOpen={() => setUserMenuOpened(true)}
      withinPortal
      trapFocus={false}
    >
      <Menu.Target>
        <Button
          variant="subtle"
          fullWidth
          justify="space-between"
          px="2xs"
          leftSection={
            <Group justify="center">
              <Avatar
                src={defaultAvatar}
                alt={"Avatar"}
                radius="xl"
                size={32}
              />
            </Group>
          }
          rightSection={
            <Group justify="flex-end">
              <IconChevronRight size={18} />
            </Group>
          }
        >
          <Stack gap={4}>
            <Text size="xs" truncate="end" ta="start">
              Name
            </Text>
            <Text size="12px" c="dimmed" truncate="end">
              username@exame.com
            </Text>
          </Stack>
        </Button>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Item>
          <Group gap="xs">
            <Avatar radius="xl" src={defaultAvatar} />
            <Stack gap={2}>
              <Group justify="space-between">
                <Text fz="sm" truncate="end">
                  Name
                </Text>
              </Group>
              <Text size="xs" c="dimmed" truncate="end" maw={200}>
                username@example.com
              </Text>
            </Stack>
          </Group>
        </Menu.Item>

        <Menu.Divider />
        <Menu.Label>Settings</Menu.Label>
        <Menu.Item
          rightSection={
            colorScheme === "light" ? (
              <IconMoon size="0.9rem" />
            ) : (
              <IconSun size="0.9rem" />
            )
          }
          onClick={() => {
            setColorScheme(colorScheme === "light" ? "dark" : "light");
          }}
        >
          {colorScheme === "light" ? "Dark mode" : "Light mode"}
        </Menu.Item>
        <Menu.Item rightSection={<IconSettings size="0.9rem" stroke={1.5} />}>
          Account settings
        </Menu.Item>

        <Menu.Divider />

        <Menu.Item
          rightSection={<IconLogout size="0.9rem" stroke={1.5} />}
          onClick={() => {}}
        >
          Logout
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
