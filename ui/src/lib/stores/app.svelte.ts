import { MODELS, type Model, type Project, type UserSettings, type GitHubRepo } from '$lib/types';
import { authAPI, projectsAPI, settingsAPI, githubAPI } from '$lib/api';

// Reactive app state using Svelte 5 runes
class AppState {
	// UI State
	modalOpen = $state(false);
	modalTaskId = $state<string | null>(null);
	settingsOpen = $state(false);
	projectMenuOpen = $state(false);
	inputProjectMenuOpen = $state(false);

	// Toast
	toastOpen = $state(false);
	toastMsg = $state('');

	// Project State
	activeProjectId = $state('');
	activeProjectName = $state('');
	projects = $state<Project[]>([]);
	repos = $state<GitHubRepo[]>([]);

	// Model
	activeModelId = $state('');

	// Voice Recording
	isRecording = $state(false);
	isTranscribing = $state(false);
	audioLevels = $state<number[]>(Array(12).fill(0));
	recordedDuration = $state(0);

	// PWA
	deferredPrompt = $state<BeforeInstallPromptEvent | null>(null);
	canInstallPWA = $state(false);

	// Auth
	isAuthenticated = $state(false);
	userEmail = $state('');
	githubConnected = $state(false);
	githubLogin = $state('');
	needsGitHubAuth = $state(false);

	// Settings
	settings = $state<UserSettings | null>(null);

	constructor() {
		if (typeof window !== 'undefined') {
			this.activeProjectId = localStorage.getItem('counterspell_active_project_id') || '';
			this.activeProjectName = localStorage.getItem('counterspell_active_project_name') || '';
			this.activeModelId = localStorage.getItem('counterspell_model') || MODELS[0].id;
		}
	}

	// ==================== INITIALIZATION ====================

	async init() {
		// Load auth status
		await this.checkAuth();
		// Load projects
		await this.loadProjects();
		// Load repos
		await this.loadRepos();
		// Load settings
		await this.loadSettings();
	}

	async checkAuth() {
		try {
			const session = await authAPI.checkSession();
			this.isAuthenticated = session.authenticated;
			this.userEmail = session.email || '';
			this.githubConnected = session.githubConnected;
			this.githubLogin = session.githubLogin || '';
			this.needsGitHubAuth = session.needsGitHubAuth;
		} catch (err) {
			console.error('Auth check failed:', err);
			this.isAuthenticated = false;
			this.githubConnected = false;
			this.needsGitHubAuth = false;
		}
	}

	async loadProjects() {
		try {
			this.projects = await projectsAPI.list();
		} catch (err) {
			console.error('Failed to load projects:', err);
		}
	}

	async loadRepos() {
		try {
			this.repos = await githubAPI.listRepos();
		} catch (err) {
			console.error('Failed to load repos:', err);
		}
	}

	async loadSettings() {
		try {
			this.settings = await settingsAPI.get();
		} catch (err) {
			console.error('Failed to load settings:', err);
		}
	}

	// ==================== GETTERS ====================

	get modelName(): string {
		const m = MODELS.find((m) => m.id === this.activeModelId);
		return m ? m.name.split(' ')[1] : this.activeModelId.split('#')[1];
	}

	// ==================== ACTIONS ====================

	async setActiveProject(id: string, name: string) {
		// If it's a repo ID (number as string), activate it first
		if (id.match(/^\d+$/)) {
			const repo = this.repos.find((r) => r.id.toString() === id);
			if (repo) {
				try {
					await projectsAPI.activate(repo.owner, repo.name);
					// After activation, we need to reload projects to get the actual project ID
					await this.loadProjects();
					const project = this.projects.find((p) => p.name === repo.full_name);
					if (project) {
						id = project.id;
						name = project.name;
					}
				} catch (err) {
					console.error('Failed to activate project:', err);
					this.showToast('Failed to activate project');
					return;
				}
			}
		}

		this.activeProjectId = id;
		this.activeProjectName = name;
		localStorage.setItem('counterspell_active_project_id', id);
		localStorage.setItem('counterspell_active_project_name', name);
		this.inputProjectMenuOpen = false;
		this.projectMenuOpen = false;
	}

	setModel(id: string) {
		this.activeModelId = id;
		localStorage.setItem('counterspell_model', id);
	}

	showToast(msg: string) {
		this.toastMsg = msg;
		this.toastOpen = true;
		setTimeout(() => {
			this.toastOpen = false;
		}, 3000);
	}

	closeModal() {
		if (!this.modalOpen) return;
		this.modalOpen = false;
		this.modalTaskId = null;
		if (history.state?.modal) {
			history.back();
		}
	}

	openModal(taskId: string) {
		this.modalTaskId = taskId;
		this.modalOpen = true;
		history.pushState({ modal: true }, '');
	}

	installPWA() {
		if (!this.deferredPrompt) return;
		this.deferredPrompt.prompt();
		this.deferredPrompt.userChoice.then((choiceResult) => {
			if (choiceResult.outcome === 'accepted') {
				this.showToast('Installing app...');
			}
			this.deferredPrompt = null;
			this.canInstallPWA = false;
		});
	}

	clearState() {
		// Reset UI state
		this.modalOpen = false;
		this.modalTaskId = null;
		this.settingsOpen = false;
		this.projectMenuOpen = false;
		this.inputProjectMenuOpen = false;

		// Reset Project State
		this.activeProjectId = '';
		this.activeProjectName = '';
		this.projects = [];
		this.repos = [];

		// Reset Auth
		this.isAuthenticated = false;
		this.userEmail = '';
		this.githubConnected = false;
		this.githubLogin = '';
		this.needsGitHubAuth = false;

		// Reset Settings
		this.settings = null;

		// Clear local storage
		if (typeof window !== 'undefined') {
			localStorage.removeItem('counterspell_active_project_id');
			localStorage.removeItem('counterspell_active_project_name');
			localStorage.removeItem('counterspell_model');
			// Clear any other app-specific keys if they exist
			sessionStorage.clear();
		}
	}

	// ==================== AUTH ACTIONS ====================

	async login() {
		await authAPI.loginWithGitHub();
	}

	async connectGitHub() {
		await authAPI.connectGitHub();
	}

	async logout() {
		this.clearState();
		await authAPI.logout();
	}

	async disconnect() {
		const confirmed = confirm(
			'Are you sure you want to disconnect GitHub and DELETE all project data? This cannot be undone.'
		);
		if (!confirmed) return;

		this.clearState();
		try {
			await authAPI.disconnect();
		} catch (err) {
			console.error('Failed to disconnect:', err);
			this.showToast('Failed to disconnect properly');
		}
	}

	// ==================== SETTINGS ACTIONS ====================

	openSettings() {
		this.settingsOpen = true;
	}

	closeSettings() {
		this.settingsOpen = false;
	}

	async saveSettings(newSettings: UserSettings) {
		try {
			await settingsAPI.save(newSettings);
			this.settings = newSettings;
			this.closeSettings();
			this.showToast('Settings saved');
		} catch (err) {
			console.error('Failed to save settings:', err);
			this.showToast('Failed to save settings');
		}
	}
}

export const appState = new AppState();

// PWA event types
interface BeforeInstallPromptEvent extends Event {
	prompt(): Promise<void>;
	userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}
