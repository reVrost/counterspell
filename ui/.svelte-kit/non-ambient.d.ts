
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	export interface AppTypes {
		RouteId(): "/" | "/app" | "/auth" | "/auth/callback" | "/dashboard" | "/disconnect" | "/login" | "/logout";
		RouteParams(): {
			
		};
		LayoutParams(): {
			"/": Record<string, never>;
			"/app": Record<string, never>;
			"/auth": Record<string, never>;
			"/auth/callback": Record<string, never>;
			"/dashboard": Record<string, never>;
			"/disconnect": Record<string, never>;
			"/login": Record<string, never>;
			"/logout": Record<string, never>
		};
		Pathname(): "/" | "/app" | "/app/" | "/auth" | "/auth/" | "/auth/callback" | "/auth/callback/" | "/dashboard" | "/dashboard/" | "/disconnect" | "/disconnect/" | "/login" | "/login/" | "/logout" | "/logout/";
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): string & {};
	}
}