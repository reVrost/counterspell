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
		// Scalability Mock Data
		"android": {ID: "android", Name: "acme/android-app", Icon: "fa-android", Color: "text-green-500"},
		"api":     {ID: "api", Name: "acme/public-api", Icon: "fa-network-wired", Color: "text-yellow-400"},
		"docs":    {ID: "docs", Name: "acme/documentation", Icon: "fa-book", Color: "text-gray-400"},
		"infra":   {ID: "infra", Name: "acme/infrastructure", Icon: "fa-network-wired", Color: "text-red-400"},
		"utils":   {ID: "utils", Name: "acme/go-utils", Icon: "fa-toolbox", Color: "text-blue-300"},
		"design":  {ID: "design", Name: "acme/design-system", Icon: "fa-paint-brush", Color: "text-pink-400"},
		"auth":    {ID: "auth", Name: "acme/auth-service", Icon: "fa-lock", Color: "text-yellow-600"},
		"cli":     {ID: "cli", Name: "acme/cli-tool", Icon: "fa-terminal", Color: "text-gray-200"},
		"billing": {ID: "billing", Name: "acme/billing-engine", Icon: "fa-credit-card", Color: "text-green-300"},
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
	http.HandleFunc("/feed", handleFeed)              // Full Feed
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
	mu.Lock()
	data := struct{ Projects map[string]Project }{Projects: projects}
	mu.Unlock()
	tmpl := template.Must(template.New("index").Parse(shellTemplate))
	tmpl.Execute(w, data)
}

func handleFeed(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	projectFilter := r.URL.Query().Get("project")

	var activeTasks []*Task
	var reviewTasks []*Task
	var doneTasks []*Task

	for _, t := range tasks {
		if projectFilter != "" && t.ProjectID != projectFilter {
			continue
		}

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
		Reviews  []*Task
		Active   []*Task
		Done     []*Task
		Projects map[string]Project
		Filter   string
	}{
		Reviews:  reviewTasks,
		Active:   activeTasks,
		Done:     doneTasks,
		Projects: projects,
		Filter:   projectFilter,
	}

	tmpl := template.New("feed").Funcs(funcMap)
	template.Must(tmpl.Parse(feedTemplate))
	template.Must(tmpl.New("activeRows").Parse(activeRowsTemplate))
	template.Must(tmpl.New("reviewsSection").Parse(reviewsTemplate))
	tmpl.Execute(w, data)
}

// Partial endpoint for just the active rows (Fixes recursion bug)
func handleFeedActive(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	var activeTasks []*Task
	// Fetch active tasks
	for _, t := range tasks {
		if t.State == StateWorking || t.State == StateQueued {
			activeTasks = append(activeTasks, t)
		}
	}
	// Sort by ID desc
	sort.Slice(activeTasks, func(i, j int) bool { return activeTasks[i].ID > activeTasks[j].ID })

	// Also fetch reviews for OOB update
	var reviewTasks []*Task
	for _, t := range tasks {
		if t.State == StateReview {
			reviewTasks = append(reviewTasks, t)
		}
	}
	sort.Slice(reviewTasks, func(i, j int) bool { return reviewTasks[i].ID > reviewTasks[j].ID })

	mu.Unlock()

	tmpl := template.New("activeRows").Funcs(funcMap)
	template.Must(tmpl.Parse(activeRowsTemplate))
	template.Must(tmpl.New("reviewsSection").Parse(reviewsTemplate))

	// Execute active rows (main swap)
	tmpl.Execute(w, activeTasks)

	// Execute reviews (OOB swap)
	w.Write([]byte(`<div id="reviews-container" hx-swap-oob="true">`))
	data := struct{ Reviews []*Task }{Reviews: reviewTasks}
	tmpl.ExecuteTemplate(w, "reviewsSection", data)
	w.Write([]byte(`</div>`))
}

func handleProjectFilter(w http.ResponseWriter, r *http.Request) {
	handleFeed(w, r)
}

func handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	task := tasks[id]
	data := struct {
		Task    *Task
		Project Project
	}{
		Task:    task,
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
	w.Header().Set("HX-Trigger", `{"closeModal": true, "toast": "Task restarting..."}`)
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
	w.Header().Set("HX-Trigger", `{"closeModal": true, "toast": "Feedback sent to agent"}`)
	handleFeed(w, r)
}

func handleActionMerge(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	t := tasks[id]
	t.State = StateApproved
	t.Logs = append(t.Logs, LogEntry{time.Now(), "Merged to main", "success"})
	mu.Unlock()

	w.Header().Set("HX-Trigger", `{"closeModal": true, "toast": "Changes merged successfully"}`)
	handleFeed(w, r)
}

func handleActionDiscard(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.URL.Path)
	mu.Lock()
	delete(tasks, id)
	mu.Unlock()

	w.Header().Set("HX-Trigger", `{"closeModal": true, "toast": "Task discarded"}`)
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
		Log      string
	}{
		{10, "Reading codebase context..."},
		{30, "Identified relevant files"},
		{50, "Drafting changes..."},
		{70, "Running unit tests..."},
		{100, "Ready for review"},
	}

	for _, p := range phases {
		// Slowing down to 1.5s per step
		time.Sleep(1000 * time.Millisecond)
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
	"hasPrefix":  strings.HasPrefix,
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
    toastMsg: '',
    toastOpen: false,

    // Onboarding State
    showOnboarding: !localStorage.getItem('conductor_v2_onboarded'),
    onboardingStep: 0, // 0: idle, 1: connecting, 2: syncing, 3: done

    startOnboarding() {
        this.onboardingStep = 1;
        // Simulate Auth
        setTimeout(() => {
            this.onboardingStep = 2;
            // Simulate Repo Sync
            setTimeout(() => {
                this.onboardingStep = 3;
                // Reveal App
                setTimeout(() => {
                    localStorage.setItem('conductor_v2_onboarded', 'true');
                    this.showOnboarding = false;
                }, 800);
            }, 1200);
        }, 1500);
    },

    showToast(msg) {
        this.toastMsg = msg;
        this.toastOpen = true;
        setTimeout(() => this.toastOpen = false, 3000);
    },
    simulateVoice() {
        this.listening = true;
        setTimeout(() => {
            this.listening = false;
            this.$refs.voiceForm.requestSubmit();
        }, 1200);
    },
    closeModal() { this.modalOpen = false; setTimeout(() => { document.getElementById('modal-content').innerHTML = ''; }, 300); }
}"
@keydown.escape="closeModal()"
@toast.window="showToast($event.detail.value)"
class="h-screen flex flex-col overflow-hidden no-tap-highlight bg-[#0C0E12]">

    <!-- Onboarding Overlay -->
    <div x-show="showOnboarding"
         x-transition:leave="transition ease-in duration-500"
         x-transition:leave-start="opacity-100 translate-y-0"
         x-transition:leave-end="opacity-0 -translate-y-10"
         class="fixed inset-0 z-[100] bg-[#0C0E12] flex flex-col items-center justify-center text-center px-6">

         <!-- Background Effects -->
         <div class="absolute inset-0 overflow-hidden pointer-events-none">
             <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"></div>
             <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animation-delay-2000 animate-pulse"></div>
         </div>

         <!-- Content -->
         <div class="relative z-10 max-w-md w-full space-y-8">
             <div class="space-y-4">
                 <div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl mx-auto flex items-center justify-center shadow-lg shadow-blue-500/20 mb-6">
                    <i class="fas fa-wave-square text-2xl text-white"></i>
                 </div>
                 <h1 class="text-3xl font-bold text-white tracking-tight">Welcome to Conductor</h1>
                 <p class="text-gray-400 text-sm leading-relaxed">
                     The AI-native orchestration layer for your engineering team.<br>
                     Connect your identity to begin.
                 </p>
             </div>

             <!-- Step 0: Initial Action -->
             <div x-show="onboardingStep === 0" x-transition.opacity>
                 <button @click="startOnboarding()"
                     class="w-full bg-white text-black font-bold h-12 rounded-lg hover:bg-gray-200 transition active:scale-95 flex items-center justify-center gap-2">
                     <i class="fab fa-github text-lg"></i> Continue with GitHub
                 </button>
                 <p class="mt-4 text-[10px] text-gray-600">By continuing, you agree to the Developer Protocol v2.1</p>
             </div>

             <!-- Step 1-3: Loading Sequence -->
             <div x-show="onboardingStep > 0" class="space-y-4" x-cloak>
                 <div class="bg-gray-900/50 rounded-xl p-4 border border-gray-800 text-left space-y-3 font-mono text-xs">

                     <!-- Item 1: Auth -->
                     <div class="flex items-center gap-3">
                         <div class="w-4 h-4 rounded-full flex items-center justify-center"
                              :class="onboardingStep > 1 ? 'bg-green-500/20 text-green-500' : 'bg-purple-500/20 text-purple-400'">
                             <i class="fas" :class="onboardingStep > 1 ? 'fa-check' : 'fa-circle-notch fa-spin'"></i>
                         </div>
                         <span :class="onboardingStep > 1 ? 'text-gray-400' : 'text-gray-200'">Authenticating with GitHub...</span>
                     </div>

                     <!-- Item 2: Repos -->
                     <div class="flex items-center gap-3" x-show="onboardingStep >= 2" x-transition.opacity>
                         <div class="w-4 h-4 rounded-full flex items-center justify-center"
                              :class="onboardingStep > 2 ? 'bg-green-500/20 text-green-500' : 'bg-blue-500/20 text-blue-400'">
                             <i class="fas" :class="onboardingStep > 2 ? 'fa-check' : 'fa-circle-notch fa-spin'"></i>
                         </div>
                         <span :class="onboardingStep > 2 ? 'text-gray-400' : 'text-gray-200'">Indexing 62 repositories...</span>
                     </div>

                     <!-- Item 3: Voice -->
                     <div class="flex items-center gap-3" x-show="onboardingStep >= 3" x-transition.opacity>
                         <div class="w-4 h-4 rounded-full bg-green-500/20 text-green-500 flex items-center justify-center">
                             <i class="fas fa-check"></i>
                         </div>
                         <span class="text-green-400">Environment Ready</span>
                     </div>
                 </div>
             </div>
         </div>
    </div>

    <!-- Toast Notification -->
    <div x-show="toastOpen"
         x-transition:enter="transition ease-out duration-300"
         x-transition:enter-start="translate-y-full opacity-0"
         x-transition:enter-end="translate-y-0 opacity-100"
         x-transition:leave="transition ease-in duration-200"
         x-transition:leave-start="translate-y-0 opacity-100"
         x-transition:leave-end="translate-y-full opacity-0"
         class="fixed top-6 left-1/2 -translate-x-1/2 z-[60] bg-gray-900 border border-gray-700/50 text-white px-4 py-2 rounded-full shadow-2xl flex items-center gap-3 text-sm font-medium">
         <i class="fas fa-check-circle text-green-500"></i>
         <span x-text="toastMsg"></span>
    </div>

    <!-- Header -->
    <header class="h-14 border-b border-linear-border bg-gray-900/80 backdrop-blur-md flex items-center justify-between px-4 z-20 shrink-0">
        <div @click="projectMenuOpen = !projectMenuOpen" class="flex items-center gap-2 cursor-pointer active:opacity-70 transition">
             <div class="w-5 h-5 rounded bg-gray-800 flex items-center justify-center text-[10px] text-gray-400 border border-gray-700">
                <i class="fas fa-layer-group"></i>
             </div>
             <span class="font-semibold text-sm tracking-tight text-gray-200">All Projects</span>
             <i class="fas fa-chevron-down text-[10px] text-gray-600"></i>
        </div>
        <div class="relative" x-data="{ userMenuOpen: false }">
             <div @click="userMenuOpen = !userMenuOpen" class="flex items-center gap-3 cursor-pointer hover:opacity-80 transition p-1">
                  <div class="h-2 w-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.4)]"></div>
                  <div class="w-6 h-6 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[10px] font-bold text-gray-300">BOB</div>
             </div>

             <div x-show="userMenuOpen" @click.outside="userMenuOpen = false" x-cloak
                  x-transition.scale.origin.top.right.duration.200ms
                  class="absolute top-10 right-0 z-50 w-48 bg-[#16191F] border border-gray-700 rounded-xl shadow-2xl overflow-hidden py-1 transform">

                  <div class="px-4 py-3 border-b border-gray-800 mb-1">
                     <p class="text-[10px] text-gray-500 uppercase tracking-wider font-bold">Current User</p>
                     <div class="flex items-center gap-2 mt-1">
                         <div class="w-4 h-4 rounded-full bg-gray-700 flex items-center justify-center text-[8px]">BOB</div>
                         <p class="text-xs font-medium text-gray-200">Bob Engineer</p>
                     </div>
                  </div>

                  <div class="px-2">
                     <div class="px-2 py-1.5 hover:bg-white/5 rounded cursor-pointer text-xs text-gray-400 flex items-center gap-2">
                         <i class="fas fa-cog w-4"></i> Settings
                     </div>
                  </div>

                  <div class="h-px bg-gray-800 my-1 mx-2"></div>

                  <div class="px-2 pb-1">
                      <div @click="localStorage.removeItem('conductor_v2_onboarded'); window.location.reload()"
                           class="px-2 py-1.5 hover:bg-red-500/10 rounded cursor-pointer text-xs text-red-400 hover:text-red-300 flex items-center gap-2 transition-colors">
                           <i class="fas fa-sign-out-alt w-4"></i> Sign Out
                      </div>
                  </div>
             </div>
        </div>
    </header>

    <!-- Project Filter Menu -->
    <div x-show="projectMenuOpen" @click.outside="projectMenuOpen = false"
         x-data="{ search: '' }"
         x-transition.opacity.duration.150ms
         class="absolute top-16 left-4 z-30 w-72 bg-[#16191F] border border-gray-700 rounded-xl shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col">

         <!-- Sticky Search Header -->
         <div class="p-3 border-b border-gray-700 bg-[#16191F]">
            <div class="relative">
                <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 text-xs"></i>
                <input x-model="search" type="text" placeholder="Filter repositories..."
                       class="w-full bg-gray-900 border border-gray-700 rounded-lg pl-8 pr-3 py-1.5 text-xs text-white focus:outline-none focus:border-blue-500 placeholder-gray-600">
            </div>
         </div>

         <!-- Scrollable List -->
         <div class="max-h-[320px] overflow-y-auto py-1">
             <div hx-get="/feed" hx-target="#feed-container"
                  @click="projectMenuOpen = false"
                  class="px-4 py-2 hover:bg-white/5 cursor-pointer text-sm font-bold text-white border-b border-gray-800/50 mb-1"
                  x-show="'all projects'.includes(search.toLowerCase())">
                  All Projects
             </div>

             {{ range .Projects }}
             <div hx-get="/feed?project={{.ID}}" hx-target="#feed-container"
                  @click="projectMenuOpen = false"
                  class="px-4 py-2 hover:bg-white/5 cursor-pointer flex items-center gap-3 group transition"
                  x-show="'{{.Name}}'.toLowerCase().includes(search.toLowerCase())">
                  <div class="w-6 h-6 rounded bg-gray-800 border border-gray-700 flex items-center justify-center shrink-0">
                    <i class="fas {{.Icon}} {{.Color}} text-[10px]"></i>
                  </div>
                  <div class="flex-1 min-w-0">
                      <div class="text-sm text-gray-400 group-hover:text-white truncate transition">{{.Name}}</div>
                  </div>
             </div>
             {{ end }}

             <!-- Empty State -->
             <div x-show="$el.parentElement.querySelectorAll('div[hx-get]:not([style*=\'display: none\'])').length === 0"
                  class="px-4 py-8 text-center text-gray-600 text-xs">
                  No projects found.
             </div>
         </div>

         <!-- Footer -->
         <div class="px-3 py-2 bg-gray-900/50 border-t border-gray-800 text-[10px] text-gray-500 flex justify-between">
            <span>{{ len .Projects }} Repositories</span>
            <span class="hover:text-blue-400 cursor-pointer"><i class="fas fa-plus"></i> New</span>
         </div>
    </div>

    <!-- Main Feed -->
    <main class="flex-1 overflow-y-auto bg-[#0C0E12] relative"
          id="feed-container"
          hx-get="/feed"
          hx-trigger="load, closeModal from:body">
    </main>

    <!-- Floating Mic -->
    <div class="fixed bottom-6 right-6 z-10">
        <form x-ref="voiceForm" hx-post="/add-task" hx-target="#feed-container" hx-swap="innerHTML">
            <input type="hidden" name="voice_input" value="Refactor the payment gateway wrapper">
            <button type="button" @click="simulateVoice()"
                :class="listening ? 'w-16 h-16 bg-purple-500 scale-110' : 'w-14 h-14 bg-white scale-100'"
                class="rounded-full shadow-[0_0_40px_rgba(168,85,247,0.3)] flex items-center justify-center text-black text-xl transition-all duration-300 active:scale-95">
                <i class="fas" :class="listening ? 'fa-wave-square animate-pulse text-white' : 'fa-microphone'"></i>
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
<div class="px-3 pt-4 pb-24">

    <!-- NEEDS REVIEW (Target for OOB Swaps) -->
    <div id="reviews-container">
        {{ template "reviewsSection" . }}
    </div>

    <!-- IN PROGRESS (Polled separately to avoid nesting bug) -->
    <div class="mb-6">
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
        <div class="space-y-2 opacity-60 hover:opacity-100 transition">
            {{ range .Done }}
            {{ $p := getProject .ProjectID }}
            <div class="bg-[#13151A] border border-gray-800/20 rounded-xl p-3 flex justify-between items-center group cursor-pointer hover:bg-gray-800/50 transition"
                 @click="modalOpen = true"
                 hx-get="/task/{{.ID}}" hx-target="#modal-content">
                <div class="flex items-center gap-3">
                    <div class="w-5 h-5 rounded-full bg-green-900/40 text-green-500 flex items-center justify-center text-[10px]">
                        <i class="fas fa-check"></i>
                    </div>
                    <div>
                         <div class="text-xs text-gray-400 line-through decoration-gray-600 group-hover:no-underline group-hover:text-gray-300 transition">{{.Description}}</div>
                         <div class="text-[10px] text-gray-600">{{$p.Name}}</div>
                    </div>
                </div>
                <i class="fas fa-chevron-right text-[10px] text-gray-700 opacity-0 group-hover:opacity-100 transition"></i>
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
    <div class="bg-[#13151A] border border-gray-800/50 rounded-xl p-3 opacity-90 transition hover:opacity-100 shadow-sm"
            @click="modalOpen = true"
            hx-get="/task/{{.ID}}" hx-target="#modal-content">
        <div class="flex justify-between items-start mb-2">
            <div class="flex items-center gap-2">
                <span class="{{$p.Color}} opacity-70 w-5 h-5 rounded flex items-center justify-center text-[10px]">
                    <i class="fas {{$p.Icon}}"></i>
                </span>
                <span class="text-xs text-gray-500">{{$p.Name}}</span>
            </div>
            <span class="text-[10px] text-purple-400 flex items-center gap-1">
                <i class="fas fa-circle-notch fa-spin"></i> {{.Progress}}%
            </span>
        </div>
        <p class="text-sm text-gray-400 leading-tight">{{.Description}}</p>

        {{ if lt .Progress 100 }}
        <div class="mt-2 w-full bg-gray-800 h-0.5 rounded-full overflow-hidden">
            <div class="bg-purple-600 h-full transition-all duration-300" style="width: {{.Progress}}%"></div>
        </div>
        {{ end }}
    </div>
{{ end }}
{{ if not . }}
    <div class="px-2 text-xs text-gray-600 italic">No active agents running...</div>
{{ end }}
`

const reviewsTemplate = `
    {{ if .Reviews }}
    <div class="mb-6">
        <h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-2">Needs Review</h3>
        <div class="space-y-2">
            {{ range .Reviews }}
            {{ $p := getProject .ProjectID }}
            <div class="bg-[#13151A] border border-gray-800 rounded-xl p-3 active:bg-gray-800 transition shadow-sm relative group"
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
        <div x-show="activeTab === 'preview'" class="p-4 flex flex-col h-full pb-32"
             x-data="{ booting: true, url: '' }"
             x-init="setTimeout(() => { booting = false; url = 'https://agent-' + Math.floor(Math.random()*1000) + '.ngrok.io' }, 1500)">

            <div x-show="booting" class="flex-1 flex flex-col items-center justify-center space-y-4 text-gray-500">
                <div class="relative">
                    <div class="w-12 h-12 rounded-full border-2 border-gray-700 border-t-purple-500 animate-spin"></div>
                    <div class="absolute inset-0 flex items-center justify-center">
                        <i class="fas fa-terminal text-[10px]"></i>
                    </div>
                </div>
                <p class="text-xs font-mono animate-pulse">Starting dev server...</p>
            </div>

            <div x-show="!booting" class="flex flex-col h-full bg-white rounded overflow-hidden">
                <!-- Mock Browser Chrome -->
                <div class="bg-gray-100 border-b border-gray-300 px-3 py-2 flex items-center gap-2 shrink-0">
                    <div class="flex gap-1.5">
                        <div class="w-2.5 h-2.5 rounded-full bg-red-400"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-yellow-400"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-green-400"></div>
                    </div>
                    <div class="ml-2 flex-1 bg-white border border-gray-300 rounded px-2 py-0.5 text-[10px] text-gray-500 font-mono flex items-center gap-1">
                        <i class="fas fa-lock text-[8px]"></i>
                        <span x-text="url"></span>
                    </div>
                </div>
                <!-- Mock Content -->
                <div class="flex-1 flex items-center justify-center bg-gray-50 relative overflow-hidden">
                     {{ if .Task.PreviewURL }}
                        <img src="{{.Task.PreviewURL}}" class="w-full h-full object-cover opacity-90 hover:opacity-100 transition">
                     {{ else }}
                        <div class="text-center">
                            <h1 class="text-2xl font-bold text-gray-900 mb-2">Conductor Interface</h1>
                            <p class="text-gray-500 text-sm">Preview build successful.</p>
                            <button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded text-xs font-bold hover:bg-blue-700">Action</button>
                        </div>
                     {{ end }}
                </div>
            </div>
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
                     <button type="submit" class="bg-purple-600 hover:bg-purple-500 text-white text-xs font-bold px-4 py-2 rounded-lg transition-colors">Send Feedback</button>
                </div>
            </form>
        </div>

        <!-- Default Buttons -->
        <div x-show="!showChat" class="grid grid-cols-4 gap-2 h-12">
            <!-- Retry Button -->
            <button hx-post="/action/retry/{{.Task.ID}}" hx-swap="none"
                class="col-span-1 bg-[#21262d] hover:bg-[#30363d] border border-gray-700/50 rounded-lg flex flex-col items-center justify-center gap-0.5 active:scale-95 transition-all text-gray-400 hover:text-white">
                <i class="fas fa-undo text-xs mb-0.5"></i>
                <span class="text-[10px] font-semibold uppercase tracking-wide">Retry</span>
            </button>

            <!-- Chat Button -->
            <button @click="showChat = true"
                class="col-span-1 bg-[#21262d] hover:bg-[#30363d] border border-gray-700/50 rounded-lg flex flex-col items-center justify-center gap-0.5 active:scale-95 transition-all text-purple-400 hover:text-purple-300">
                <i class="fas fa-sparkles text-xs mb-0.5"></i>
                <span class="text-[10px] font-semibold uppercase tracking-wide">Chat</span>
            </button>

            <!-- Merge Button -->
            <button hx-post="/action/merge/{{.Task.ID}}" hx-swap="none"
                class="col-span-2 bg-white hover:bg-gray-100 text-black rounded-lg flex items-center justify-center gap-2 font-bold text-sm active:scale-95 transition-all">
                <i class="fas fa-code-branch"></i>
                <span>Merge</span>
            </button>
        </div>
    </div>
</div>
`
