import { AppShell } from "@mantine/core";
import { NavBar } from "./components/Navbar";
import { Route, Routes } from "react-router-dom";
import MetricsPage from "./pages/Metrics";
import LogsPage from "./pages/Logs";

function App() {
  return (
    <AppShell
      navbar={{
        width: 180,
        breakpoint: "sm",
      }}
    >
      <AppShell.Navbar>
        <NavBar />
      </AppShell.Navbar>
      <AppShell.Main>
        <Routes>
          <Route path="/microscope" element={<LogsPage />} />
          <Route path="/microscope/logs" element={<LogsPage />} />
          <Route path="/microscope/metrics" element={<MetricsPage />} />
          <Route path="/microscope/*" element={<LogsPage />} />
        </Routes>
      </AppShell.Main>
    </AppShell>
  );
}

export default App;
