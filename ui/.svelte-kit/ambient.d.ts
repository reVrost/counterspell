
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```sh
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const MANPATH: string;
	export const LDFLAGS: string;
	export const __MISE_DIFF: string;
	export const NODE: string;
	export const INIT_CWD: string;
	export const SHELL: string;
	export const TERM: string;
	export const CPPFLAGS: string;
	export const HOMEBREW_REPOSITORY: string;
	export const TMPDIR: string;
	export const npm_config_global_prefix: string;
	export const GOBIN: string;
	export const DIRENV_DIR: string;
	export const WINDOWID: string;
	export const COLOR: string;
	export const npm_config_noproxy: string;
	export const npm_config_local_prefix: string;
	export const USER: string;
	export const N_PREFIX: string;
	export const COMMAND_MODE: string;
	export const OPENAI_API_KEY: string;
	export const npm_config_globalconfig: string;
	export const SSH_AUTH_SOCK: string;
	export const SUPABASE_JWT_SECRET: string;
	export const SUPABASE_ANON_KEY: string;
	export const __CF_USER_TEXT_ENCODING: string;
	export const npm_execpath: string;
	export const VIRTUAL_ENV_DISABLE_PROMPT: string;
	export const MULTI_TENANT: string;
	export const DIRENV_WATCHES: string;
	export const PATH: string;
	export const npm_package_json: string;
	export const LaunchInstanceID: string;
	export const _: string;
	export const GITHUB_REDIRECT_URI: string;
	export const npm_config_userconfig: string;
	export const npm_config_init_module: string;
	export const __CFBundleIdentifier: string;
	export const PWD: string;
	export const npm_command: string;
	export const OPENROUTER_API_KEY: string;
	export const npm_lifecycle_event: string;
	export const EDITOR: string;
	export const KITTY_PID: string;
	export const npm_package_name: string;
	export const LANG: string;
	export const npm_config_npm_version: string;
	export const XPC_FLAGS: string;
	export const SUPABASE_URL: string;
	export const npm_config_node_gyp: string;
	export const npm_package_version: string;
	export const DIRENV_FILE: string;
	export const XPC_SERVICE_NAME: string;
	export const GEMINI_API_KEY: string;
	export const HOME: string;
	export const SHLVL: string;
	export const TERMINFO: string;
	export const SERPER_API_KEY: string;
	export const __MISE_ORIG_PATH: string;
	export const GOROOT: string;
	export const HOMEBREW_PREFIX: string;
	export const MISE_SHELL: string;
	export const npm_config_cache: string;
	export const LOGNAME: string;
	export const npm_lifecycle_script: string;
	export const GITHUB_CLIENT_SECRET: string;
	export const PKG_CONFIG_PATH: string;
	export const npm_config_user_agent: string;
	export const KITTY_INSTALLATION_DIR: string;
	export const KITTY_WINDOW_ID: string;
	export const PROMPT_EOL_MARK: string;
	export const HOMEBREW_CELLAR: string;
	export const INFOPATH: string;
	export const __MISE_SESSION: string;
	export const KITTY_LISTEN_ON: string;
	export const CONDA_CHANGEPS1: string;
	export const DIRENV_DIFF: string;
	export const SECURITYSESSIONID: string;
	export const npm_node_execpath: string;
	export const npm_config_prefix: string;
	export const GITHUB_CLIENT_ID: string;
	export const KITTY_PUBLIC_KEY: string;
	export const COLORTERM: string;
	export const NODE_ENV: string;
}

/**
 * Similar to [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		MANPATH: string;
		LDFLAGS: string;
		__MISE_DIFF: string;
		NODE: string;
		INIT_CWD: string;
		SHELL: string;
		TERM: string;
		CPPFLAGS: string;
		HOMEBREW_REPOSITORY: string;
		TMPDIR: string;
		npm_config_global_prefix: string;
		GOBIN: string;
		DIRENV_DIR: string;
		WINDOWID: string;
		COLOR: string;
		npm_config_noproxy: string;
		npm_config_local_prefix: string;
		USER: string;
		N_PREFIX: string;
		COMMAND_MODE: string;
		OPENAI_API_KEY: string;
		npm_config_globalconfig: string;
		SSH_AUTH_SOCK: string;
		SUPABASE_JWT_SECRET: string;
		SUPABASE_ANON_KEY: string;
		__CF_USER_TEXT_ENCODING: string;
		npm_execpath: string;
		VIRTUAL_ENV_DISABLE_PROMPT: string;
		MULTI_TENANT: string;
		DIRENV_WATCHES: string;
		PATH: string;
		npm_package_json: string;
		LaunchInstanceID: string;
		_: string;
		GITHUB_REDIRECT_URI: string;
		npm_config_userconfig: string;
		npm_config_init_module: string;
		__CFBundleIdentifier: string;
		PWD: string;
		npm_command: string;
		OPENROUTER_API_KEY: string;
		npm_lifecycle_event: string;
		EDITOR: string;
		KITTY_PID: string;
		npm_package_name: string;
		LANG: string;
		npm_config_npm_version: string;
		XPC_FLAGS: string;
		SUPABASE_URL: string;
		npm_config_node_gyp: string;
		npm_package_version: string;
		DIRENV_FILE: string;
		XPC_SERVICE_NAME: string;
		GEMINI_API_KEY: string;
		HOME: string;
		SHLVL: string;
		TERMINFO: string;
		SERPER_API_KEY: string;
		__MISE_ORIG_PATH: string;
		GOROOT: string;
		HOMEBREW_PREFIX: string;
		MISE_SHELL: string;
		npm_config_cache: string;
		LOGNAME: string;
		npm_lifecycle_script: string;
		GITHUB_CLIENT_SECRET: string;
		PKG_CONFIG_PATH: string;
		npm_config_user_agent: string;
		KITTY_INSTALLATION_DIR: string;
		KITTY_WINDOW_ID: string;
		PROMPT_EOL_MARK: string;
		HOMEBREW_CELLAR: string;
		INFOPATH: string;
		__MISE_SESSION: string;
		KITTY_LISTEN_ON: string;
		CONDA_CHANGEPS1: string;
		DIRENV_DIFF: string;
		SECURITYSESSIONID: string;
		npm_node_execpath: string;
		npm_config_prefix: string;
		GITHUB_CLIENT_ID: string;
		KITTY_PUBLIC_KEY: string;
		COLORTERM: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
