import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

interface User {
	login: string;
	id: number;
	avatar_url: string;
	name: string | null;
}

interface AuthState {
	token: string | null;
	user: User | null;
	loading: boolean;
}

function createAuthStore() {
	const initialState: AuthState = {
		token: browser ? localStorage.getItem('github_token') : null,
		user: browser ? JSON.parse(localStorage.getItem('github_user') || 'null') : null,
		loading: false
	};

	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,
		setToken: (token: string) => {
			if (browser) {
				localStorage.setItem('github_token', token);
			}
			update((state) => ({ ...state, token }));
		},
		setUser: (user: User) => {
			if (browser) {
				localStorage.setItem('github_user', JSON.stringify(user));
			}
			update((state) => ({ ...state, user }));
		},
		setLoading: (loading: boolean) => {
			update((state) => ({ ...state, loading }));
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('github_token');
				localStorage.removeItem('github_user');
			}
			set({ token: null, user: null, loading: false });
		},
		initialize: async () => {
			update((state) => ({ ...state, loading: true }));
			
			const token = browser ? localStorage.getItem('github_token') : null;
			if (!token) {
				update((state) => ({ ...state, loading: false }));
				return;
			}

			try {
				const response = await fetch('https://api.github.com/user', {
					headers: {
						Authorization: `Bearer ${token}`,
						Accept: 'application/vnd.github+json'
					}
				});

				if (response.ok) {
					const user = await response.json();
					if (browser) {
						localStorage.setItem('github_user', JSON.stringify(user));
					}
					update((state) => ({ ...state, user, loading: false }));
				} else {
					// Token invalid, clear it
					if (browser) {
						localStorage.removeItem('github_token');
						localStorage.removeItem('github_user');
					}
					set({ token: null, user: null, loading: false });
				}
			} catch (error) {
				console.error('Failed to fetch user:', error);
				update((state) => ({ ...state, loading: false }));
			}
		}
	};
}

export const auth = createAuthStore();
export const isAuthenticated = derived(auth, ($auth) => !!$auth.token && !!$auth.user);
