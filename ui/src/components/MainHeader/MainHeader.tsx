import { Group, Title } from "@mantine/core";
import { IconLayoutSidebarLeftCollapse } from "@tabler/icons-react";
import { useLocation } from "react-router-dom";

const getTitle = (pathname: string) => {
  switch (pathname) {
    case "/home":
      return "Home";
    case "/agents":
      return "Agents";
    case "/logs":
      return "Logs";
    case "/metrics":
      return "Metrics";
    case "/settings":
      return "Settings";
    default:
      return "Home";
  }
};

export const MainHeader = () => {
  const location = useLocation();
  const title = getTitle(location.pathname);

  return (
    <Group
      align="center"
      style={{
        borderBottom: "1px solid var(--mantine-color-gray-2)",
      }}
      p="xs"
      gap="xs"
    >
      <IconLayoutSidebarLeftCollapse size={20} stroke={1.5} />
      <Title order={6} fw={500}>
        {title}
      </Title>
    </Group>
  );
};

