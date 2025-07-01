import { AppShell } from "@mantine/core";
import { NavBar } from "./components/Navbar";
import { Route, Routes } from "react-router-dom";
import LogsPage from "./pages/Logs";
import MetricsPage from "./pages/Metrics";
import SettingsPage from "./pages/Settings";

function App() {
  return (
    <AppShell
      navbar={{ width: 180, breakpoint: "sm", collapsed: { mobile: true } }}
    >
      <AppShell.Navbar>
        <NavBar />
      </AppShell.Navbar>
      <AppShell.Main>
        <Routes>
          <Route path="/counterspell/logs" element={<LogsPage />} />
          <Route path="/counterspell/metrics" element={<MetricsPage />} />
          <Route path="/counterspell/settings" element={<SettingsPage />} />
        </Routes>
      </AppShell.Main>
    </AppShell>
  );
}

export default App;
