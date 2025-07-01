import {
  IconChartLine,
  IconGauge,
  IconHome2,
  IconNotes,
  IconSettings,
} from "@tabler/icons-react";
import { Stack, Text, Title, Group } from "@mantine/core";
import classes from "./Navbar.module.css";
import { NavLink } from "react-router-dom";

interface NavbarLinkProps {
  icon: typeof IconHome2;
  label: string;
  path: string;
}

function NavbarLink({ icon: Icon, label, path }: NavbarLinkProps) {
  return (
    <NavLink
      className={classes.navlink}
      to={"counterspell/" + path}
      style={({ isActive }) => {
        return {
          color: isActive
            ? "var(--mantine-color-black)"
            : "var(--mantine-color-white)",
          backgroundColor: isActive
            ? "var(--mantine-color-white)"
            : "var(--mantine-color-black)",
          boxShadow: isActive ? "var(--mantine-shadow-sm)" : "",
        };
      }}
    >
      <Icon size={20} stroke={1.5} />
      <Text fw="500" size="sm" ml={8}>
        {label}
      </Text>
    </NavLink>
  );
}

const menu = [
  { icon: IconNotes, label: "Logs", path: "logs" },
  { icon: IconGauge, label: "Metrics", path: "metrics" },
  { icon: IconSettings, label: "Settings", path: "settings" },
];

export function NavBar() {
  const links = menu.map((link) => <NavbarLink {...link} key={link.label} />);

  return (
    <nav className={classes.navbar}>
      <div>
        <Group mb="30" gap={0}>
          <Title order={1}>M</Title>
          <IconChartLine size={40} stroke={1.5} />
        </Group>
        <Stack
          justify="flex-start"
          align="flex-start"
          gap="lg"
          className={classes.navbarMain}
        >
          {links}
        </Stack>
      </div>
    </nav>
  );
}
