<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { GitHubClient, GITHUB_CLIENT_ID } from '$lib/github';

	let error = '';
	let status = 'Authenticating...';

	onMount(async () => {
		const code = $page.url.searchParams.get('code');
		
		if (!code) {
			error = 'No authorization code received';
			return;
		}

		try {
			status = 'Exchanging code for token...';
			
			// Note: In production, this should go through a backend to keep client_secret secure
			// For now, we'll use a proxy approach or the user needs to set up a backend
			// This is a simplified version that assumes a token is returned directly
			
			// Exchange code for token via proxy/backend
			// For development, you can use a service like https://github-oauth-proxy.vercel.app
			// or set up your own backend endpoint
			
			const response = await fetch('/api/auth/callback', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ code })
			});

			if (!response.ok) {
				throw new Error('Failed to exchange code for token');
			}

			const { access_token } = await response.json();
			
			status = 'Getting user info...';
			auth.setToken(access_token);

			// Get user info
			const userResponse = await fetch('https://api.github.com/user', {
				headers: {
					Authorization: `Bearer ${access_token}`,
					Accept: 'application/vnd.github+json'
				}
			});

			if (!userResponse.ok) {
				throw new Error('Failed to get user info');
			}

			const user = await userResponse.json();
			auth.setUser(user);

			status = 'Checking for pact repo...';
			
			// Check if user has pact repo, create if not
			const github = new GitHubClient(access_token);
			const repoExists = await github.repoExists(user.login);
			
			if (!repoExists) {
				status = 'Creating your pact repo...';
				await github.createRepo();
				
				// Wait a moment for GitHub to initialize
				await new Promise(resolve => setTimeout(resolve, 2000));
				
				// Create initial pact.json
				status = 'Setting up initial configuration...';
				const initialConfig = {
					version: '1.0.0',
					user: user.login,
					modules: {
						shell: {},
						editor: {},
						git: {},
						ai: { providers: {}, prompts: {}, agents: {} },
						tools: { configs: {} }
					},
					secrets: []
				};
				
				await github.savePactConfig(user.login, initialConfig);
			}

			status = 'Redirecting to dashboard...';
			goto('/dashboard');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Authentication failed';
			console.error('Auth error:', e);
		}
	});
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono flex items-center justify-center">
	<div class="text-center space-y-4">
		{#if error}
			<div class="text-red-400">
				<p class="text-xl font-bold">Authentication Error</p>
				<p class="text-sm mt-2">{error}</p>
				<a href="/" class="text-emerald-400 hover:text-emerald-300 text-sm mt-4 block">
					Back to home
				</a>
			</div>
		{:else}
			<div class="animate-spin w-8 h-8 border-2 border-emerald-400 border-t-transparent rounded-full mx-auto"></div>
			<p class="text-zinc-400">{status}</p>
		{/if}
	</div>
</div>
