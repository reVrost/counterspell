import {
  IconGauge,
  IconHome,
  IconHome2,
  IconNotes,
  IconSettings,
} from "@tabler/icons-react";
import { Anchor, Button, Group, Stack, Text } from "@mantine/core";
import classes from "./Navbar.module.css";
import { UserButton } from "./UserButton/UserButton";
import { TeamButton } from "./TeamButton/TeamButton";

interface NavbarLinkProps {
  icon: typeof IconHome2;
  label: string;
  path: string;
}

function NavbarLink({ icon: Icon, label, path }: NavbarLinkProps) {
  return (
    <Anchor href={"#/" + path} underline="never" w="100%">
      <Group flex={1} justify="flex-start">
        <Button
          fullWidth
          variant="subtle"
          justify="flex-start"
          leftSection={<Icon size={20} stroke={1.5} />}
          size="xs"
        >
          <Text fw="500" size="sm" ta="start">
            {label}
          </Text>
        </Button>
      </Group>
    </Anchor>
  );
}

const menu = [
  { icon: IconHome, label: "Home", path: "logs" },
  { icon: IconNotes, label: "Logs", path: "logs" },
  { icon: IconGauge, label: "Metrics", path: "metrics" },
  { icon: IconSettings, label: "Settings", path: "settings" },
];

export function NavBar() {
  const links = menu.map((link) => <NavbarLink {...link} key={link.label} />);

  return (
    <nav className={classes.navbar}>
      <TeamButton />
      <Stack mt="lg" gap="2xs" justify="flex-start" align="flex-start">
        {links}
      </Stack>
      <Stack mt="auto">
        <UserButton />
      </Stack>
    </nav>
  );
}
