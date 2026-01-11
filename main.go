package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// --- DATA MODELS ---

type Project struct {
	ID    string
	Name  string
	Icon  string // FontAwesome class
	Color string // Tailwind color class stub
}

type LogEntry struct {
	Timestamp time.Time
	Message   string
	Type      string // "info", "error", "success", "agent"
}

type TaskState string

const (
	StateQueued   TaskState = "queued"
	StateWorking  TaskState = "working"
	StateReview   TaskState = "review"
	StateApproved TaskState = "approved"
	StateRejected TaskState = "rejected"
)

type Task struct {
	ID          int
	ProjectID   string
	Description string
	AgentName   string
	Progress    int
	State       TaskState
	Summary     string
	MockDiff    string
	Logs        []LogEntry
	CreatedAt   time.Time
	PreviewURL  string // Mock preview
}

// --- STORE ---

var (
	projects = map[string]Project{
		"core": {ID: "core", Name: "acme/core-platform", Icon: "fa-server", Color: "text-blue-400"},
		"web":  {ID: "web", Name: "acme/web-dashboard", Icon: "fa-columns", Color: "text-purple-400"},
		"ios":  {ID: "ios", Name: "acme/ios-app", Icon: "fa-mobile-alt", Color: "text-green-400"},
	}

	tasks  = make(map[int]*Task)
	nextID = 1
	mu     sync.Mutex
)

// --- HANDLERS ---

func main() {
	rand.Seed(time.Now().UnixNano())
	initStore()

	// Static/HTMX Routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/feed", handleFeed)           // Full Feed
	http.HandleFunc("/feed/active", handleFeedActive) // Partial for Polling
	http.HandleFunc("/project-filter", handleProjectFilter)
	
	// Task Logic
	http.HandleFunc("/add-task", handleAddTask)
	http.HandleFunc("/task/", handleTaskDetail) // The "Workbench" view
	
	// Actions
	http.HandleFunc("/action/retry/", handleActionRetry)
	http.HandleFunc("/action/merge/", handleActionMerge)
	http.HandleFunc("/action/chat/", handleActionChat)
	http.HandleFunc("/action/discard/", handleActionDiscard)

	fmt.Println("---------------------------------------------------------")
	fmt.Println("conductor v2.1 // Flow Fixes")
	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("---------------------------------------------------------")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initStore() {
	mu.Lock()
	defer mu.Unlock()
	
	// Completed Task
	createTask("Refactor auth middleware", "core")
	t1 := tasks[1]
	t1.State = StateApproved
	t1.Logs = append(t1.Logs, LogEntry{time.Now(), "Merged to main", "success"})
	
	// Review Task
	createTask("Add Dark Mode to Dashboard", "web")
	t2 := tasks[2]
	t2.State = StateReview
	t2.Progress = 100
	t2.PreviewURL = "https://cdn.dribbble.com/users/1615584/screenshots/15710288/media/7845f7478d59d56223253b8b603d1544.jpg?resize=400x300&vertical=center" // abstract img
	t2.MockDiff = `file: src/App.css
@@ -10,2 +10,3 @@
 :root {
-  --bg: #fff;
+  --bg: #111;
+  --text: #eee;
 }`
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(shellTemplate))
	tmpl.Execute(w, nil)
}

func handleFeed(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	projectFilter := r.URL.Query().Get("project")
	
	var activeTasks []*Task
	var reviewTasks []*Task
	var doneTasks   []*Task
	
	for _, t := range tasks {
		if projectFilter != "" && t.ProjectID != projectFilter { continue }
		
		if t.State == StateReview {
			reviewTasks = append(reviewTasks, t)
		} else if t.State == StateWorking || t.State == StateQueued {
			activeTasks = append(activeTasks, t)
		} else if t.State == StateApproved {
			doneTasks = append(doneTasks, t)
		}
	}
	
	// Sort by ID desc
	sort.Slice(reviewTasks, func(i, j int) bool { return reviewTasks[i].ID > reviewTasks[j].ID })
	sort.Slice(activeTasks, func(i, j int) bool { return activeTasks[i].ID > activeTasks[j].ID })
	sort.Slice(doneTasks, func(i, j int) bool { return doneTasks[i].ID > doneTasks[j].ID })
	
	mu.Unlock()

	data := struct {
		Reviews []*Task
		Active  []*Task
		Done    []*Task
		Projects map[string]Project
		Filter   string
	}{
		Reviews: reviewTasks,
		Active:  activeTasks,
		Done:    doneTasks,
		Projects: projects,
		Filter:  projectFilter,
	}

	tmpl := template.Must(template.New("feed").Funcs(funcMap).Parse(feedTemplate))
	tmpl.Execute(w, data)
}

// Partial endpoint for just the active rows (Fixes recursion bug)
func handleFeedActive(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	var activeTasks []*Task
	for _, t := range tasks {
		if t.State == StateWorking || t.State == StateQueued {
			activeTasks = append(activeTasks, t)
		}
	}
	sort.Slice(activeTasks, func(i, j int) bool { return activeTasks[i].ID > activeTasks[j].ID })
	mu.Unlock()

	tmpl := template.Must(template.New("activeRows").Funcs(funcMap).Parse(activeRowsTemplate))
	tmpl.Execute(w, activeTasks)
}

func handleProjectFilter(w http.ResponseWriter, r *http.Request) {
	handleFeed(w, r)
}

func handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	task := tasks[id]
	data := struct{
		Task *Task
		Project Project
	}{
		Task: task,
		Project: projects[task.ProjectID],
	}
	mu.Unlock()

	tmpl := template.Must(template.New("detail").Funcs(funcMap).Parse(detailTemplate))
	tmpl.Execute(w, data)
}

func handleAddTask(w http.ResponseWriter, r *http.Request) {
	voiceText := r.FormValue("voice_input")
	if voiceText == "" {
		voiceText = "Update the user schema to support OIDC."
	}
	
	pIDs := []string{"core", "web", "ios"}
	pID := pIDs[rand.Intn(len(pIDs))]

	mu.Lock()
	t := createTask(voiceText, pID)
	mu.Unlock()
	
	go startAgentWork(t.ID)
	handleFeed(w, r)
}

// Actions

func handleActionRetry(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	t := tasks[id]
	t.State = StateWorking
	t.Progress = 0
	t.Logs = append(t.Logs, LogEntry{time.Now(), "User requested retry", "info"})
	mu.Unlock()
	
	go startAgentWork(id)
	w.Header().Set("HX-Trigger", "closeModal")
	handleFeed(w, r)
}

func handleActionChat(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	msg := r.FormValue("message")
	
	mu.Lock()
	t := tasks[id]
	t.State = StateWorking
	t.Progress = 0
	t.Logs = append(t.Logs, LogEntry{time.Now(), "Refinement: " + msg, "info"})
	mu.Unlock()

	go startAgentWork(id)
	w.Header().Set("HX-Trigger", "closeModal")
	handleFeed(w, r)
}

func handleActionMerge(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	t := tasks[id]
	t.State = StateApproved
	t.Logs = append(t.Logs, LogEntry{time.Now(), "Merged to main", "success"})
	mu.Unlock()
	
	w.Header().Set("HX-Trigger", "closeModal")
	handleFeed(w, r)
}

func handleActionDiscard(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	delete(tasks, id)
	mu.Unlock()
	
	w.Header().Set("HX-Trigger", "closeModal")
	handleFeed(w, r)
}

// --- HELPERS ---

func createTask(desc, projectID string) *Task {
	id := nextID
	nextID++
	
	t := &Task{
		ID:          id,
		ProjectID:   projectID,
		Description: desc,
		AgentName:   fmt.Sprintf("Agent-%03d", rand.Intn(999)),
		Progress:    0,
		State:       StateQueued,
		CreatedAt:   time.Now(),
		Summary:     "Analyzing dependency graph...",
		MockDiff:    `// Loading diff...`,
		Logs: []LogEntry{
			{time.Now(), "Task queued via voice", "info"},
		},
	}
	tasks[id] = t
	return t
}

func startAgentWork(id int) {
	phases := []struct {
		Progress int
		Log string
	}{
		{10, "Reading codebase context..."},
		{30, "Identified relevant files"},
		{50, "Drafting changes..."},
		{70, "Running unit tests..."},
		{100, "Ready for review"},
	}

	for _, p := range phases {
		time.Sleep(300 * time.Millisecond)
		mu.Lock()
		if t, ok := tasks[id]; ok {
			t.Progress = p.Progress
			t.Logs = append(t.Logs, LogEntry{time.Now(), p.Log, "agent"})
			if p.Progress == 100 {
				t.State = StateReview
				generateMockDiff(t)
			}
		}
		mu.Unlock()
	}
}

func generateMockDiff(t *Task) {
	t.Summary = "Updated implementation."
	t.MockDiff = `file: pkg/server.go
@@ -12,4 +12,5 @@
+ // Optimization for mobile
  func HandleRequest(r *Request) {
-    process(r)
+    go processAsync(r)
  }`
}

func parseID(path string) int {
	var id int
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		fmt.Sscanf(parts[len(parts)-1], "%d", &id)
	}
	return id
}

var funcMap = template.FuncMap{
	"splitLines": func(s string) []string { return strings.Split(s, "\n") },
	"hasPrefix": strings.HasPrefix,
	"getProject": func(id string) Project { return projects[id] },
}

// --- TEMPLATES ---

const shellTemplate = `
<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>Conductor</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                fontFamily: { sans: ['Inter', 'system-ui', 'sans-serif'], mono: ['JetBrains Mono', 'Menlo', 'monospace'] },
                extend: {
                    colors: {
                        gray: { 850: '#1A1D24', 900: '#111318', 950: '#0C0E12' },
                        linear: { border: '#2E323A', text: '#EBEBEB', sub: '#8A8F98' }
                    }
                }
            }
        }
    </script>
    <style>
        body { background-color: #0C0E12; color: #EBEBEB; -webkit-font-smoothing: antialiased; }
        .hide-scrollbar::-webkit-scrollbar { display: none; }
        .no-tap-highlight { -webkit-tap-highlight-color: transparent; }
        
        /* Slide-over animation */
        .slide-enter-active, .slide-leave-active { transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
        .slide-enter-start, .slide-leave-end { transform: translateY(100%); }
        .slide-enter-end, .slide-leave-start { transform: translateY(0); }
        
        /* Custom syntax highlight feel */
        .code-block { font-family: 'JetBrains Mono', monospace; font-size: 13px; line-height: 1.5; }
        .diff-add { background-color: rgba(46, 160, 67, 0.15); display: block; }
        .diff-del { background-color: rgba(248, 81, 73, 0.15); display: block; }
    </style>
</head>
<body x-data="{
    modalOpen: false,
    activeTab: 'diff',
    projectMenuOpen: false,
    listening: false,
    simulateVoice() {
        this.listening = true;
        setTimeout(() => {
            this.listening = false;
            document.querySelector('form button[type=\'button\']').closest('form').requestSubmit();
        }, 1200);
    },
    closeModal() { this.modalOpen = false; setTimeout(() => { document.getElementById('modal-content').innerHTML = ''; }, 300); }
}"
@keydown.escape="closeModal()"
class="h-screen flex flex-col overflow-hidden no-tap-highlight">

    <!-- Header -->
    <header class="h-14 border-b border-linear-border bg-gray-900/80 backdrop-blur-md flex items-center justify-between px-4 z-20 shrink-0">
        <div @click="projectMenuOpen = !projectMenuOpen" class="flex items-center gap-2 cursor-pointer active:opacity-70 transition">
             <div class="w-5 h-5 rounded bg-gray-800 flex items-center justify-center text-[10px] text-gray-400 border border-gray-700">
                <i class="fas fa-layer-group"></i>
             </div>
             <span class="font-semibold text-sm tracking-tight text-gray-200">All Projects</span>
             <i class="fas fa-chevron-down text-[10px] text-gray-600"></i>
        </div>
        <div class="flex items-center gap-3">
             <div class="h-2 w-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.4)]"></div>
             <div class="w-6 h-6 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[10px]">BOB</div>
        </div>
    </header>

    <!-- Project Filter Menu -->
    <div x-show="projectMenuOpen" @click.outside="projectMenuOpen = false"
         x-transition.opacity.duration.200ms
         class="absolute top-16 left-4 z-30 w-48 bg-gray-900 border border-linear-border rounded-xl shadow-2xl overflow-hidden py-1">
         <div hx-get="/feed" hx-target="#feed-container" @click="projectMenuOpen = false" class="px-4 py-3 hover:bg-white/5 cursor-pointer text-sm font-medium border-b border-gray-800">All Projects</div>
         <div hx-get="/feed?project=core" hx-target="#feed-container" @click="projectMenuOpen = false" class="px-4 py-2 hover:bg-white/5 cursor-pointer text-sm text-gray-400 hover:text-white flex items-center gap-2"><i class="fas fa-server text-blue-400 w-4"></i> Core</div>
         <div hx-get="/feed?project=web" hx-target="#feed-container" @click="projectMenuOpen = false" class="px-4 py-2 hover:bg-white/5 cursor-pointer text-sm text-gray-400 hover:text-white flex items-center gap-2"><i class="fas fa-columns text-purple-400 w-4"></i> Web</div>
    </div>

    <!-- Main Feed -->
    <main class="flex-1 overflow-y-auto bg-[#0C0E12] relative" 
          id="feed-container"
          hx-get="/feed" 
          hx-trigger="load, closeModal from:body">
    </main>

    <!-- Floating Mic -->
    <div class="fixed bottom-6 right-6 z-10">
        <form hx-post="/add-task" hx-target="#feed-container" hx-swap="innerHTML">
            <input type="hidden" name="voice_input" value="Refactor the payment gateway wrapper">
            <button type="button" @click="simulateVoice()"
                :class="listening ? 'w-16 h-16 bg-red-500 scale-110' : 'w-14 h-14 bg-white scale-100'"
                class="rounded-full shadow-[0_0_40px_rgba(255,255,255,0.15)] flex items-center justify-center text-black text-xl transition-all duration-300 active:scale-95">
                <i class="fas" :class="listening ? 'fa-wave-square animate-pulse' : 'fa-microphone'"></i>
            </button>
        </form>
    </div>

    <!-- Detail Modal -->
    <div x-show="modalOpen" class="fixed inset-0 z-40 bg-gray-950/40 backdrop-blur-[2px]" x-transition.opacity></div>
    <div x-show="modalOpen"
         class="fixed inset-x-0 bottom-0 top-[40px] z-50 bg-[#16191F] rounded-t-[24px] shadow-2xl flex flex-col overflow-hidden border-t border-white/10"
         x-transition:enter="transition transform duration-300 ease-out"
         x-transition:enter-start="translate-y-full"
         x-transition:enter-end="translate-y-0"
         x-transition:leave="transition transform duration-200 ease-in"
         x-transition:leave-start="translate-y-0"
         x-transition:leave-end="translate-y-full">
         
         <div class="h-6 w-full flex items-center justify-center shrink-0 cursor-pointer hover:bg-white/5 transition" @click="closeModal()">
            <div class="w-12 h-1 rounded-full bg-gray-700"></div>
         </div>
         <div id="modal-content" class="flex-1 flex flex-col h-full overflow-hidden"></div>
    </div>
</body>
</html>
`

const feedTemplate = `
<div class="px-3 pt-4 pb-24 space-y-6">

    <!-- NEEDS REVIEW -->
    {{ if .Reviews }}
    <div>
        <h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-2">Needs Review</h3>
        <div class="space-y-2">
            {{ range .Reviews }}
            {{ $p := getProject .ProjectID }}
            <div class="bg-gray-900 border border-gray-800 rounded-xl p-3 active:bg-gray-800 transition shadow-sm relative group"
                 @click="modalOpen = true"
                 hx-get="/task/{{.ID}}" hx-target="#modal-content">
                
                <div class="flex justify-between items-start mb-1">
                    <div class="flex items-center gap-2">
                        <span class="{{$p.Color}} bg-gray-800/50 border border-gray-700/50 w-5 h-5 rounded flex items-center justify-center text-[10px]">
                            <i class="fas {{$p.Icon}}"></i>
                        </span>
                        <span class="text-xs font-medium text-gray-400">{{$p.Name}}</span>
                    </div>
                    <span class="text-[10px] text-orange-400 bg-orange-400/10 px-1.5 py-0.5 rounded font-medium border border-orange-400/20">Review</span>
                </div>
                
                <p class="text-sm text-gray-200 font-medium leading-tight mb-2 pr-4">{{.Description}}</p>
                <div class="flex items-center gap-3 text-[11px] text-gray-600 font-mono">
                    <span><i class="fas fa-robot mr-1"></i>{{.AgentName}}</span>
                </div>
                <div class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-700">
                    <i class="fas fa-chevron-right text-xs"></i>
                </div>
            </div>
            {{ end }}
        </div>
    </div>
    {{ end }}

    <!-- IN PROGRESS (Polled separately to avoid nesting bug) -->
    <div>
        <h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-2">In Progress</h3>
        <div class="space-y-2" hx-get="/feed/active" hx-trigger="every 2s" hx-swap="innerHTML">
            <!-- Initial load -->
            {{ template "activeRows" .Active }}
        </div>
    </div>

    <!-- DONE HISTORY -->
    {{ if .Done }}
    <div class="pt-4 border-t border-gray-800/50">
        <h3 class="px-2 text-xs font-bold text-gray-600 uppercase tracking-wider mb-2">Completed</h3>
        <div class="space-y-2 opacity-60">
            {{ range .Done }}
            {{ $p := getProject .ProjectID }}
            <div class="bg-[#13151A] border border-gray-800/20 rounded-xl p-3 flex justify-between items-center group">
                <div class="flex items-center gap-3">
                    <div class="w-5 h-5 rounded-full bg-green-900/40 text-green-500 flex items-center justify-center text-[10px]">
                        <i class="fas fa-check"></i>
                    </div>
                    <div>
                         <div class="text-xs text-gray-400 line-through decoration-gray-600">{{.Description}}</div>
                         <div class="text-[10px] text-gray-600">{{$p.Name}}</div>
                    </div>
                </div>
            </div>
            {{ end }}
        </div>
    </div>
    {{ end }}
</div>
`

// Just the rows for polling
const activeRowsTemplate = `
{{ range . }}
    {{ $p := getProject .ProjectID }}
    <div class="bg-[#13151A] border border-gray-800/50 rounded-xl p-3 opacity-90 transition hover:opacity-100"
            @click="modalOpen = true"
            hx-get="/task/{{.ID}}" hx-target="#modal-content">
        <div class="flex justify-between items-start mb-2">
            <div class="flex items-center gap-2">
                <span class="{{$p.Color}} opacity-70 w-5 h-5 rounded flex items-center justify-center text-[10px]">
                    <i class="fas {{$p.Icon}}"></i>
                </span>
                <span class="text-xs text-gray-500">{{$p.Name}}</span>
            </div>
            <span class="text-[10px] text-blue-400 flex items-center gap-1">
                <i class="fas fa-circle-notch fa-spin"></i> {{.Progress}}%
            </span>
        </div>
        <p class="text-sm text-gray-400 leading-tight">{{.Description}}</p>
        
        {{ if lt .Progress 100 }}
        <div class="mt-2 w-full bg-gray-800 h-0.5 rounded-full overflow-hidden">
            <div class="bg-blue-600 h-full transition-all duration-300" style="width: {{.Progress}}%"></div>
        </div>
        {{ end }}
    </div>
{{ end }}
{{ if not . }}
    <div class="px-2 text-xs text-gray-600 italic">No active agents running...</div>
{{ end }}
`

const detailTemplate = `
<div class="flex flex-col h-full" x-data="{ showChat: false, activeTab: 'diff' }">
    <!-- Modal Header -->
    <div class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-[#16191F]">
        <div class="flex items-center gap-3">
             <button @click="closeModal()" class="w-8 h-8 rounded-full hover:bg-white/5 flex items-center justify-center text-gray-400">
                <i class="fas fa-arrow-left"></i>
             </button>
             <div>
                <div class="flex items-center gap-2">
                    <span class="{{.Project.Color}} text-[10px]"><i class="fas {{.Project.Icon}}"></i> {{.Project.Name}}</span>
                    <span class="text-[10px] text-gray-600 font-mono">#{{.Task.ID}}</span>
                </div>
                <h2 class="text-sm font-bold text-gray-200 line-clamp-1 w-48">{{.Task.Description}}</h2>
             </div>
        </div>
        
        <!-- Trash Action -->
        <button hx-post="/action/discard/{{.Task.ID}}" hx-swap="none" 
            class="text-gray-600 hover:text-red-400 transition" onclick="if(!confirm('Discard task?')) return false;">
            <i class="fas fa-trash"></i>
        </button>
    </div>
    
    <!-- Tabs Container -->
    <div class="flex items-center justify-center p-2 bg-[#16191F] shrink-0 border-b border-white/5">
        <div class="flex bg-gray-900 rounded-lg p-0.5 border border-gray-700/50">
             <button @click="activeTab = 'diff'" 
                :class="activeTab === 'diff' ? 'bg-gray-800 text-white shadow' : 'text-gray-500'"
                class="px-4 py-1 text-[11px] font-medium rounded-md transition-all">Code</button>
             <button @click="activeTab = 'preview'" 
                :class="activeTab === 'preview' ? 'bg-gray-800 text-white shadow' : 'text-gray-500'"
                class="px-4 py-1 text-[11px] font-medium rounded-md transition-all">Preview</button>
             <button @click="activeTab = 'activity'" 
                :class="activeTab === 'activity' ? 'bg-gray-800 text-white shadow' : 'text-gray-500'"
                class="px-4 py-1 text-[11px] font-medium rounded-md transition-all">Log</button>
        </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 overflow-y-auto bg-[#0D1117] relative w-full">
        
        <!-- Tab 1: DIFF -->
        <div x-show="activeTab === 'diff'" class="p-0 min-h-full pb-32">
            {{ if eq .Task.State "working" }}
                <div class="flex flex-col items-center justify-center h-64 text-gray-500 space-y-4">
                    <i class="fas fa-cog fa-spin text-3xl opacity-50"></i>
                    <p class="text-xs font-mono">Generating changes...</p>
                </div>
            {{ else }}
                <div class="px-4 py-3 border-b border-gray-800 sticky top-0 bg-[#0D1117] z-10 flex justify-between">
                    <span class="text-xs text-gray-400 font-mono">pkg/server.go</span>
                    <span class="text-[10px] text-green-500 font-mono">+12 / -2</span>
                </div>
                <div class="p-4 font-mono text-xs leading-6 text-gray-300 whitespace-pre overflow-x-auto">{{range splitLines .Task.MockDiff}}<div class="{{if hasPrefix . "+"}}diff-add{{else if hasPrefix . "-"}}diff-del{{end}}">{{.}}</div>{{end}}</div>
            {{ end }}
        </div>

        <!-- Tab 2: PREVIEW -->
        <div x-show="activeTab === 'preview'" class="p-4 flex items-center justify-center h-full pb-32">
            {{ if .Task.PreviewURL }}
                <img src="{{.Task.PreviewURL}}" class="rounded-lg border border-gray-700 shadow-2xl max-w-full">
            {{ else }}
                <div class="text-center text-gray-600">
                    <i class="fas fa-eye-slash text-4xl mb-3 opacity-50"></i>
                    <p class="text-xs">No visual preview available.</p>
                </div>
            {{ end }}
        </div>

        <!-- Tab 3: ACTIVITY -->
        <div x-show="activeTab === 'activity'" class="p-5 pb-32 space-y-6">
            <div class="relative border-l border-gray-800 ml-2 space-y-6">
                {{ range .Task.Logs }}
                <div class="ml-4 relative">
                    <div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] 
                        {{if eq .Type "agent"}}bg-blue-500{{else if eq .Type "success"}}bg-green-500{{else}}bg-gray-600{{end}}"></div>
                    <div class="flex justify-between items-start">
                        <span class="text-xs font-bold text-gray-300 block mb-0.5">{{.Type}}</span>
                        <span class="text-[10px] text-gray-600 font-mono">Just now</span>
                    </div>
                    <p class="text-xs text-gray-400">{{.Message}}</p>
                </div>
                {{ end }}
            </div>
        </div>
    </div>

    <!-- Bottom Actions Toolbar (Sticky) -->
    <div class="shrink-0 p-4 border-t border-white/5 bg-[#16191F] pb-8">
        
        <!-- Chat Input Mode -->
        <div x-show="showChat" x-transition.origin.bottom class="mb-2">
            <form hx-post="/action/chat/{{.Task.ID}}" hx-swap="none" @submit="showChat = false">
                <textarea name="message" class="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-sm text-white focus:outline-none focus:border-blue-500 mb-2 font-mono h-24" placeholder="Give feedback to the agent..."></textarea>
                <div class="flex justify-between">
                     <button type="button" @click="showChat = false" class="text-xs text-gray-500 px-2">Cancel</button>
                     <button type="submit" class="bg-blue-600 text-white text-xs font-bold px-4 py-2 rounded-lg">Send Feedback</button>
                </div>
            </form>
        </div>

        <!-- Default Buttons -->
        <div x-show="!showChat" class="grid grid-cols-4 gap-2 h-12">
            <!-- Ralph Wiggum Button -->
            <button hx-post="/action/retry/{{.Task.ID}}" hx-swap="none"
                class="col-span-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg flex flex-col items-center justify-center gap-1 active:scale-95 transition">
                <i class="fas fa-dice font-sm"></i>
                <span class="text-[9px] font-bold uppercase">Retry</span>
            </button>
            
            <!-- Chat Button -->
            <button @click="showChat = true" 
                class="col-span-1 bg-gray-800 hover:bg-gray-700 text-blue-400 rounded-lg flex flex-col items-center justify-center gap-1 active:scale-95 transition">
                <i class="fas fa-comment-alt font-sm"></i>
                <span class="text-[9px] font-bold uppercase">Refine</span>
            </button>
            
            <!-- Merge Button -->
            <button hx-post="/action/merge/{{.Task.ID}}" hx-swap="none"
                class="col-span-2 bg-white text-black hover:bg-gray-200 rounded-lg flex items-center justify-center gap-2 font-bold text-sm active:scale-95 transition shadow-[0_0_15px_rgba(255,255,255,0.1)]">
                <i class="fas fa-check"></i> Approve & Merge
            </button>
        </div>
    </div>
</div>
`
