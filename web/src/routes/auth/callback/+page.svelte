<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { GitHubClient } from '$lib/github';

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

			status = 'Checking your account...';
			
			// Check if user has my-pact repo
			const github = new GitHubClient(access_token);
			const repoExists = await github.repoExists(user.login);
			
			if (repoExists) {
				// Existing user, go to dashboard
				goto('/dashboard');
			} else {
				// New user, go to setup
				goto('/setup');
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Authentication failed';
			console.error('Auth error:', e);
		}
	});
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono flex items-center justify-center">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<div class="relative z-10 text-center space-y-4">
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
