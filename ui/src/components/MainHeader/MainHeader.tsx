import { Group, Title } from "@mantine/core";
import { useLocation } from "react-router-dom";

const getTitle = (pathname: string) => {
  switch (pathname) {
    case "/home":
      return "Home";
    case "/inferences":
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
      justify="space-between"
      align="center"
      style={{
        backgroundColor: "var(--mantine-color-gray-0)",
        borderBottom: "1px solid var(--mantine-color-gray-2)",
        boxShadow: "0.5px 0.5px  var(--mantine-color-gray-2)",
      }}
      p="12"
    >
      <Title order={3}>{title}</Title>
    </Group>
  );
};