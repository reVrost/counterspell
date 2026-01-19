# Counterspell 🧙‍♂️

**An auth-free, local-first AI development agent with GitHub-style UI**

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Svelte](https://img.shields.io/badge/Svelte-5.0+-ff3e00.svg)](https://svelte.dev/)
[![TailwindCSS](https://img.shields.io/badge/Tailwind-3.0+-38bdf8.svg)](https://tailwindcss.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ✨ Features

### 🎯 Core Capabilities
- **AI Task Management** - Create, track, and retry AI-powered development tasks
- **File Browser** - Browse, view, edit, and delete files with syntax highlighting
- **Git Operations** - Full Git integration with GitHub-style diffs
- **Real-Time Updates** - SSE-powered live updates for tasks
- **Auth-Free** - No authentication required, works locally
- **Dark Theme** - Beautiful Vercel-inspired dark mode

### 🛠️ Technical Features
- **Code Syntax Highlighting** - Prism.js with auto-detection for 20+ languages
- **GitHub-Style Diffs** - Side-by-side diff viewer exactly like GitHub
- **Real-Time SSE** - Server-Sent Events for instant updates
- **Responsive Design** - Works on all screen sizes
- **Type Safe** - Full TypeScript support

### 🎨 UI/UX
- Beautiful dark theme inspired by Vercel
- Smooth transitions and animations
- Loading states with spinners
- Error handling with toast messages
- Empty states with helpful messaging
- Keyboard-accessible navigation

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

## 📚 Usage Guide

### Creating Tasks

1. Navigate to the Tasks page (`/`)
2. Enter your task intent (e.g., "Add user authentication to the API")
3. Optionally set project ID and model
4. Click "Create Task"

The AI agent will:
- Analyze your request
- Plan the implementation
- Execute the plan
- Provide real-time updates

### Browsing Files

1. Navigate to the Files page (`/files`)
2. Click on directories to navigate
3. Click on files to view them
4. Use the editor to modify files
5. Click "Save" to commit changes

Files are displayed with:
- Syntax highlighting based on file type
- Language auto-detection
- GitHub-style dark theme

### Managing Git

1. Navigate to the Git page (`/git`)
2. View current branch and status
3. Stage files for commit
4. Create commits with messages
5. Manage branches (create, checkout)
6. Pull and push changes
7. Click "View Diff" to see GitHub-style diffs

### Settings

1. Navigate to Settings (`/settings`)
2. Configure API keys (OpenRouter, Zai, Anthropic, OpenAI)
3. Set agent backend (native or claude-code)
4. Click "Save Settings"

---

## 🔧 Architecture

### Backend (Go)
- **Framework**: Built with `net/http`
- **Database**: SQLite for local storage
- **Git Integration**: Native Go git operations
- **SSE**: Server-Sent Events for real-time updates
- **API**: RESTful with JSON responses

### Frontend (SvelteKit)
- **Framework**: SvelteKit 5 with Svelte 5 runes
- **Styling**: TailwindCSS with custom theme
- **State Management**: Svelte 5 reactive stores
- **Syntax Highlighting**: Prism.js
- **Icons**: Lucide Svelte

### API Endpoints

#### Tasks
- `GET /api/v1/tasks` - List all tasks
- `GET /api/v1/task/{id}` - Get task details
- `POST /api/v1/action/add` - Create task
- `POST /api/v1/action/clear/{id}` - Delete task
- `POST /api/v1/action/retry/{id}` - Retry task

#### Files
- `GET /api/v1/files/list` - List files
- `GET /api/v1/files/read` - Read file
- `POST /api/v1/files/write` - Write file
- `DELETE /api/v1/files/delete` - Delete file
- `GET /api/v1/files/search` - Search files

#### Git
- `GET /api/v1/git/status` - Get git status
- `GET /api/v1/git/branches` - List branches
- `GET /api/v1/git/log` - Get commit log
- `GET /api/v1/git/diff` - Get git diff
- `POST /api/v1/git/add` - Stage files
- `POST /api/v1/git/commit` - Create commit
- `POST /api/v1/git/checkout` - Checkout branch
- `POST /api/v1/git/branch` - Create branch
- `GET /api/v1/git/pull` - Pull changes
- `GET /api/v1/git/push` - Push changes

#### Settings
- `GET /api/v1/settings` - Get settings
- `POST /api/v1/settings` - Update settings

#### SSE
- `GET /api/v1/events` - Server-Sent Events stream

---

## 🎨 Customization

### Theme Colors
Edit `ui/src/app.css` to customize:
- Primary colors
- Background colors
- Accent colors
- Border colors

### API Configuration
Edit backend environment variables:
```bash
export OPENROUTER_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export ZAI_API_KEY="your-key"
```

### Git Configuration
The Git operations use system Git configuration. Configure in terminal:
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
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

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🎊 Acknowledgments

- **Vercel** - For the beautiful design inspiration
- **GitHub** - For the Git diff UI inspiration
- **Svelte Team** - For the amazing framework
- **Go Team** - For the excellent language
- **Prism.js Team** - For syntax highlighting

---

## 📞 Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Join our Discord community
- Email: support@counterspell.dev

---

## 🎯 Roadmap

### Completed ✅
- [x] Auth-free local-first mode
- [x] AI task management
- [x] File browser and editor
- [x] Git operations
- [x] Real-time SSE updates
- [x] Code syntax highlighting
- [x] GitHub-style diffs
- [x] Beautiful dark theme

### Future 🔮
- [ ] Multi-model support (Claude, GPT-4, etc.)
- [ ] Terminal integration
- [ ] Task templates
- [ ] Collaboration features
- [ ] Advanced Git operations (rebase, cherry-pick)
- [ ] File history viewer
- [ ] Custom themes
- [ ] Keyboard shortcuts

---

## 🌟 Star History

If you find Counterspell helpful, please consider giving it a ⭐ star on GitHub!

---

**Made with ❤️ by the Counterspell Team**

**WE REACHED THE SUMMIT OF MOUNT DOOM!** 🏔️✨
