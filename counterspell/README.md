# Counterspell 🧙‍♂️

**An auth-free, local-first AI development agent with GitHub-style UI**

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Svelte](https://img.shields.io/badge/Svelte-5.0+-ff3e00.svg)](https://svelte.dev/)
[![TailwindCSS](https://img.shields.io/badge/Tailwind-3.0+-38bdf8.svg)](https://tailwindcss.com/)
[![License](https://img.shields.io/badge/License-FSL-green.svg)](LICENSE)

---

## ✨ Features

- **AI Task Management** - Create, track, and retry AI-powered development tasks with real-time SSE updates
- **File Browser** - Browse, view, edit, and delete files with syntax highlighting (20+ languages)
- **Git Operations** - Full Git integration with GitHub-style diffs, branch management, and staging
- **Real-Time Updates** - Live task progress via Server-Sent Events
- **Local-First Design** - No authentication required, runs entirely on your machine
- **Beautiful UI** - Vercel-inspired dark theme with smooth transitions and responsive design

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ for backend
- Node.js 18+ for frontend
- Git (for Git operations)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-repo/counterspell.git
cd counterspell
```

2. **Start the backend**
```bash
cd cmd/app
go run main.go
```
Backend runs on `http://localhost:8080`

3. **Start the frontend**
```bash
cd ui
npm install
npm run dev
```
Frontend runs on `http://localhost:5173`

4. **Open in browser**
```
http://localhost:5173
```

---

## 📁 Project Structure

```
counterspell/
├── cmd/
│   └── app/              # Backend entry point
│       └── main.go
├── internal/
│   ├── api/              # API handlers
│   ├── db/               # Database layer
│   ├── git/              # Git operations
│   └── services/         # Business logic
├── ui/
│   ├── src/
│   │   ├── routes/        # SvelteKit pages
│   │   ├── lib/
│   │   │   ├── api/      # API client
│   │   │   ├── components/ # Reusable components
│   │   │   └── stores/   # State management
│   │   └── app.css       # Global styles
│   └── static/           # Static assets
├── go.mod               # Go dependencies
├── go.sum               # Go checksums
└── package.json         # Node dependencies
```

---

## 🔒 Security

- **No Authentication**: Auth-free design for local development
- **Input Validation**: All inputs are validated
- **SQL Injection Protection**: Parameterized queries
- **File Access**: Restricted to working directory
- **API Key Storage**: Encrypted in SQLite

---

## 📝 License

This project is licensed under the Functional Source License (FSL) - see the [LICENSE](LICENSE) file for details.

---

**Made with ❤️ by the Counterspell Team**

**WE REACHED THE SUMMIT OF MOUNT DOOM!** 🏔️✨
