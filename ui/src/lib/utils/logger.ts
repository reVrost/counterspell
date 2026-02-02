/**
 * UI Logger - sends logs to server for agent debugging
 *
 * All logs go to server.log with source="ui" attribute.
 * Agents can read logs via GET /api/v1/logs
 */

type LogLevel = "error" | "warn" | "info" | "debug";

interface LogEntry {
  level: LogLevel;
  message: string;
  component?: string;
  stack?: string;
  url?: string;
  extra?: Record<string, unknown>;
}

const API_BASE = import.meta.env.DEV ? "" : "";
const ENABLE_REMOTE_LOGGING = false;

async function sendLog(entry: LogEntry): Promise<void> {
  entry.url = window.location.href;

  try {
    if (!ENABLE_REMOTE_LOGGING) {
      return;
    }
    await fetch(`${API_BASE}/api/v1/log`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(entry),
      credentials: "include",
    });
  } catch {
    // Silently fail - don't cause more errors trying to log
    console.warn("[logger] Failed to send log to server");
  }
}

/**
 * Log an error to the server
 */
export function logError(
  message: string,
  options?: {
    component?: string;
    stack?: string;
    extra?: Record<string, unknown>;
  },
): void {
  return;
  console.error(`[${options?.component || "app"}]`, message);
  sendLog({
    level: "error",
    message,
    component: options?.component,
    stack: options?.stack,
    extra: options?.extra,
  });
}

/**
 * Log a warning to the server
 */
export function logWarn(
  message: string,
  options?: { component?: string; extra?: Record<string, unknown> },
): void {
  return;
  console.warn(`[${options?.component || "app"}]`, message);
  sendLog({
    level: "warn",
    message,
    component: options?.component,
    extra: options?.extra,
  });
}

/**
 * Log info to the server
 */
export function logInfo(
  message: string,
  options?: { component?: string; extra?: Record<string, unknown> },
): void {
  console.info(`[${options?.component || "app"}]`, message);
  sendLog({
    level: "info",
    message,
    component: options?.component,
    extra: options?.extra,
  });
}

/**
 * Log debug info to the server (only in dev mode locally, always sends to server)
 */
export function logDebug(
  message: string,
  options?: { component?: string; extra?: Record<string, unknown> },
): void {
  if (import.meta.env.DEV) {
    console.debug(`[${options?.component || "app"}]`, message);
  }
  sendLog({
    level: "debug",
    message,
    component: options?.component,
    extra: options?.extra,
  });
}

/**
 * Capture an Error object and log it
 */
export function captureException(
  error: Error,
  options?: { component?: string; extra?: Record<string, unknown> },
): void {
  logError(error.message, {
    component: options?.component,
    stack: error.stack,
    extra: {
      name: error.name,
      ...options?.extra,
    },
  });
}

/**
 * Initialize global error handlers - call once at app startup
 */
export function initGlobalErrorHandlers(): void {
  // Catch unhandled errors
  window.addEventListener("error", (event) => {
    logError(event.message, {
      component: "global",
      stack: event.error?.stack,
      extra: {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      },
    });
  });

  // Catch unhandled promise rejections
  window.addEventListener("unhandledrejection", (event) => {
    const message = event.reason?.message || String(event.reason);
    logError(`Unhandled rejection: ${message}`, {
      component: "global",
      stack: event.reason?.stack,
      extra: {
        reason: String(event.reason),
      },
    });
  });

  logInfo("UI logger initialized", { component: "logger" });
}
