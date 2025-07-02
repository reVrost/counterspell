import React from "react";
import ReactDOM from "react-dom/client";
import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import { ModalsProvider } from "@mantine/modals";
import { HashRouter } from "react-router-dom";
import App from "./App";
import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import "@mantine/code-highlight/styles.css";
import { shadcnCssVariableResolver } from "./cssVariableResolver.ts";
import { shadcnTheme } from "./theme.ts";
import "./style.css";

let myTheme = shadcnTheme;

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <HashRouter>
      <MantineProvider
        theme={myTheme}
        cssVariablesResolver={shadcnCssVariableResolver}
      >
        <ModalsProvider>
          <Notifications />
          <App />
        </ModalsProvider>
      </MantineProvider>
    </HashRouter>
  </React.StrictMode>,
);
