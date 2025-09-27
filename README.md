# Counterspell

Counterspell is an LLM agent execution runtime with a focus on observability. It provides a framework for defining and running LLM-powered agents, configured via simple YAML files. With its built-in web UI, you can monitor traces, logs, and interact with your agents in real-time.

**⚠️ This project is a work in progress and is not yet ready for production use. ⚠️**

## Key Features

- **LLM Agent Execution Runtime**: Define and run complex agents with different execution modes (`plan`, `loop`, `single`).
- **YAML-based Configuration**: Easily configure your agents, models, and prompts using simple YAML files.
- **Built-in Observability**: Comes with OpenTelemetry tracing and logging out-of-the-box, backed by a local SQLite database.
- **Web UI**: A comprehensive web interface for visualizing traces, inspecting logs, and interacting with your agents.
- **Go Backend, React Frontend**: A modern and performant tech stack.
- **REST and RPC APIs**: For programmatic access to the runtime and observability data.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go (1.25 or later)
- Node.js (20.x or later)

### Running with Docker Compose

The easiest way to get Counterspell up and running is with Docker Compose.

```bash
docker-compose up
```

This will start the backend server on port `8080` and the frontend UI on port `5173`.

- **Backend API**: `http://localhost:8080`
- **Frontend UI**: `http://localhost:5173`

### Development Environment

For development, you can use the provided `dev.sh` script, which uses `kitty` to create a split-pane terminal for the backend and frontend.

```bash
./dev.sh
```

## How it Works

Counterspell's core concepts are **Agents**, **Runtimes**, and **Sessions**.

- **Agents**: The basic building blocks of your LLM-powered applications. An agent is defined by its ID, model, prompt, and execution mode.
- **Runtime**: The execution environment for your agents. The runtime manages the lifecycle of agents and executes them according to their configuration.
- **Session**: Represents a single interaction with an agent, including the initial mission and any intermediate messages.

### Agent Configuration

Agents are configured in YAML files. Here's an example of a simple "writer" agent:

```yaml
agents:
  - id: "writer"
    mode: "plan"
    model: "google/gemini-2.0-flash-lite-001"
    prompt: |
      You are a creative writer.
      Your mission is to: {{.mission}}.
      Your output must be a JSON object matching this schema:
      {
        "plan": [
          {
            "kind": "agent",
            "id": "final_writer",
            "params": {}
          }
        ]
      }
  - id: "final_writer"
    model: "google/gemini-2.0-flash-lite-001"
    mode: "single"
    prompt: |
      Write a short story about {{.mission}}.

session:
  root_agent_id: "writer"
  mission: "a robot who discovers music"
```

## Architecture

Counterspell consists of three main components:

- **Backend**: A Go server built with the Echo framework. It provides the core runtime, APIs, and serves the web UI.
- **Frontend**: A React application built with Vite and Mantine for the UI.
- **Database**: A local SQLite database for storing traces and logs.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the [MIT License](LICENSE).

