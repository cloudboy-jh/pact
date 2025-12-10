// GitHub OAuth configuration
// Replace with your OAuth App credentials
export const GITHUB_CLIENT_ID = 'YOUR_GITHUB_CLIENT_ID';

const GITHUB_API = 'https://api.github.com';

export interface GitHubFile {
	name: string;
	path: string;
	sha: string;
	size: number;
	type: 'file' | 'dir';
	content?: string;
}

export interface GitHubRepo {
	name: string;
	full_name: string;
	private: boolean;
	default_branch: string;
}

export interface PactConfig {
	version: string;
	user: string;
	modules: Record<string, unknown>;
	secrets: string[];
}

// Get the OAuth authorization URL
export function getAuthUrl(): string {
	const params = new URLSearchParams({
		client_id: GITHUB_CLIENT_ID,
		scope: 'repo',
		redirect_uri: `${window.location.origin}/auth/callback`
	});
	return `https://github.com/login/oauth/authorize?${params}`;
}

// Create a class for GitHub API operations
export class GitHubClient {
	private token: string;

	constructor(token: string) {
		this.token = token;
	}

	private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
		const response = await fetch(`${GITHUB_API}${endpoint}`, {
			...options,
			headers: {
				Authorization: `Bearer ${this.token}`,
				Accept: 'application/vnd.github+json',
				'Content-Type': 'application/json',
				...options.headers
			}
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({}));
			throw new Error(error.message || `GitHub API error: ${response.status}`);
		}

		return response.json();
	}

	// Check if the user's pact repo exists
	async repoExists(username: string): Promise<boolean> {
		try {
			await this.request(`/repos/${username}/pact`);
			return true;
		} catch {
			return false;
		}
	}

	// Create the pact repo
	async createRepo(): Promise<GitHubRepo> {
		return this.request<GitHubRepo>('/user/repos', {
			method: 'POST',
			body: JSON.stringify({
				name: 'pact',
				description: 'My development environment configuration - managed by pact',
				private: false,
				auto_init: true
			})
		});
	}

	// Get repo contents
	async getContents(username: string, path = ''): Promise<GitHubFile[]> {
		const endpoint = `/repos/${username}/pact/contents/${path}`;
		const result = await this.request<GitHubFile | GitHubFile[]>(endpoint);
		return Array.isArray(result) ? result : [result];
	}

	// Get file content
	async getFileContent(username: string, path: string): Promise<string> {
		const file = await this.request<GitHubFile>(`/repos/${username}/pact/contents/${path}`);
		if (file.content) {
			return atob(file.content);
		}
		throw new Error('File has no content');
	}

	// Create or update a file
	async updateFile(
		username: string,
		path: string,
		content: string,
		message: string,
		sha?: string
	): Promise<void> {
		await this.request(`/repos/${username}/pact/contents/${path}`, {
			method: 'PUT',
			body: JSON.stringify({
				message,
				content: btoa(content),
				sha
			})
		});
	}

	// Delete a file
	async deleteFile(username: string, path: string, sha: string, message: string): Promise<void> {
		await this.request(`/repos/${username}/pact/contents/${path}`, {
			method: 'DELETE',
			body: JSON.stringify({
				message,
				sha
			})
		});
	}

	// Get pact.json
	async getPactConfig(username: string): Promise<PactConfig | null> {
		try {
			const content = await this.getFileContent(username, 'pact.json');
			return JSON.parse(content);
		} catch {
			return null;
		}
	}

	// Save pact.json
	async savePactConfig(username: string, config: PactConfig, sha?: string): Promise<void> {
		const content = JSON.stringify(config, null, 2);
		await this.updateFile(username, 'pact.json', content, 'Update pact.json', sha);
	}

	// Get the SHA of a file (needed for updates)
	async getFileSha(username: string, path: string): Promise<string | null> {
		try {
			const file = await this.request<GitHubFile>(`/repos/${username}/pact/contents/${path}`);
			return file.sha;
		} catch {
			return null;
		}
	}

	// Create directory structure by creating a placeholder file
	async createDirectory(username: string, path: string): Promise<void> {
		await this.updateFile(username, `${path}/.gitkeep`, '', `Create ${path} directory`);
	}
}
