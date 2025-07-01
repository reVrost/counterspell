import {
  IconChartLine,
  IconGauge,
  IconHome2,
  IconNotes,
  IconSettings,
} from "@tabler/icons-react";
import { Stack, UnstyledButton, Text, Title, Group } from "@mantine/core";
import classes from "./Navbar.module.css";
import { NavLink } from "react-router-dom";

interface NavbarLinkProps {
  icon: typeof IconHome2;
  label: string;
  onClick?: () => void;
  path: string;
}

function NavbarLink({ icon: Icon, label, onClick, path }: NavbarLinkProps) {
  return (
    <UnstyledButton onClick={onClick} px="sm" className={classes.link} w="100%">
      <NavLink
        className={classes.navlink}
        to={"microscope/" + path}
        style={({ isActive }) => {
          return {
            color: isActive
              ? "var(--mantine-color-white)"
              : "var(--mantine-color-white)",
            textDecoration: isActive ? "none" : "none",
          };
        }}
      >
        <Icon size={20} stroke={1.5} />
        <Text fw="500" size="sm" ml={8}>
          {label}
        </Text>
      </NavLink>
    </UnstyledButton>
  );
}

const menu = [
  // { icon: IconHome2, label: "Home" },
  { icon: IconNotes, label: "Logs", path: "logs" },
  { icon: IconGauge, label: "Metrics", path: "metrics" },
  { icon: IconSettings, label: "Settings", path: "settings" },
  // { icon: IconUser, label: "Account" },
  // { icon: IconFingerprint, label: "Security" },
];

export function NavBar() {
  const links = menu.map((link) => <NavbarLink {...link} key={link.label} />);

  return (
    <nav className={classes.navbar}>
      <div>
        <Group mb="40" gap={0}>
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

      {/* <Stack justify="center" gap={0}> */}
      {/*   <NavbarLink icon={IconSwitchHorizontal} label="Change account" /> */}
      {/*   <NavbarLink icon={IconLogout} label="Logout" /> */}
      {/* </Stack> */}
    </nav>
  );
}
