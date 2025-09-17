import { AppShell, Card, Stack } from "@mantine/core";
import { NavBar } from "./components/Navbar";
import { Route, Routes } from "react-router-dom";
import LogsPage from "./pages/Logs";
import MetricsPage from "./pages/Metrics";
import SettingsPage from "./pages/Settings";
import { SecretProvider } from "./context/SecretContext";
import { MainHeader } from "./components/MainHeader/MainHeader";
import AgentsPage from "./pages/AgentsPage";
import CreateBlueprintPage from "./pages/CreateBlueprintPage";

function App() {
  return (
    <SecretProvider>
      <AppShell
        navbar={{ width: 256, breakpoint: "sm", collapsed: { mobile: true } }}
        style={{
          backgroundColor: "var(--mantine-color-gray-0)",
        }}
      >
        <AppShell.Navbar withBorder={false}>
          <NavBar />
        </AppShell.Navbar>
        <AppShell.Main>
          <Stack>
            <Card
              h="calc(100vh - var(--mantine-spacing-lg))"
              my="2xs"
              mr="xs"
              ml="4xs"
              p={0}
              style={{ backgroundColor: "var(--mantine-color-white)" }}
              shadow="3xs"
              variant="outline"
              radius="md"
              withBorder
            >
              <MainHeader />
              <Routes>
                <Route path="/home" element={<LogsPage />} />
                <Route path="/agents" element={<AgentsPage />} />
                <Route
                  path="/agents/create"
                  element={<CreateBlueprintPage />}
                />
                <Route path="/logs" element={<LogsPage />} />
                <Route path="/metrics" element={<MetricsPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/*" element={<LogsPage />} />
              </Routes>
            </Card>
          </Stack>
        </AppShell.Main>
      </AppShell>
    </SecretProvider>
  );
}

export default App;
