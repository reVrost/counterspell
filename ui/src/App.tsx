import { AppShell } from "@mantine/core";
import { NavBar } from "./components/Navbar";
import { Route, Routes } from "react-router-dom";
import LogsPage from "./pages/Logs";
import MetricsPage from "./pages/Metrics";
import SettingsPage from "./pages/Settings";
import { SecretProvider } from "./context/SecretContext";

function App() {
  return (
    <SecretProvider>
      <AppShell
        navbar={{ width: 180, breakpoint: "sm", collapsed: { mobile: true } }}
      >
        <AppShell.Navbar>
          <NavBar />
        </AppShell.Navbar>
        <AppShell.Main>
          <Routes>
            <Route path="/logs" element={<LogsPage />} />
            <Route path="/metrics" element={<MetricsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/*" element={<LogsPage />} />
          </Routes>
        </AppShell.Main>
      </AppShell>
    </SecretProvider>
  );
}

export default App;
