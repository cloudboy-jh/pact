// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		interface Platform {
			env?: {
				GITHUB_CLIENT_SECRET?: string;
				GITHUB_CLIENT_ID?: string;
			};
		}
	}
}

// Environment variables loaded via Vite
interface ImportMetaEnv {
	readonly VITE_GITHUB_CLIENT_ID: string;
	readonly GITHUB_CLIENT_SECRET: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}

export {};
