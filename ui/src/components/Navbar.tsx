import {
  IconGauge,
  IconHome,
  IconHome2,
  IconNotes,
  IconSettings,
} from "@tabler/icons-react";
import { Button, Group, Stack, Text } from "@mantine/core";
import classes from "./Navbar.module.css";
import { UserButton } from "./UserButton/UserButton";
import { TeamButton } from "./TeamButton/TeamButton";
import { NavLink } from "react-router-dom";

interface NavbarLinkProps {
  icon: typeof IconHome2;
  label: string;
  path: string;
}

function NavbarLink({ icon: Icon, label, path }: NavbarLinkProps) {
  return (
    <NavLink
      to={"/" + path}
      style={{ textDecoration: "none", width: "100%" }}
      className={({ isActive }) => (isActive ? classes.active : "")}
    >
      <Group flex={1} justify="flex-start">
        <Button
          fullWidth
          variant="subtle"
          justify="flex-start"
          leftSection={<Icon size={20} stroke={1.5} />}
        >
          <Text size="sm" ta="start">
            {label}
          </Text>
        </Button>
      </Group>
    </NavLink>
  );
}

const menu = [
  { icon: IconHome, label: "Home", path: "home" },
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
