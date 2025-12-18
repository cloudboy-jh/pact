<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { GitHubClient } from '$lib/github';
	import { Check, Loader2, Github, Terminal, RefreshCw, Key } from 'lucide-svelte';

	type SetupState = 'checking' | 'confirm' | 'creating' | 'complete' | 'error';
	
	let state: SetupState = 'checking';
	let error = '';
	
	// Progress steps
	let steps = [
		{ id: 'repo', label: 'Creating repository', status: 'pending' as 'pending' | 'active' | 'complete' | 'error' },
		{ id: 'config', label: 'Initializing pact.json', status: 'pending' as 'pending' | 'active' | 'complete' | 'error' }
	];

	const pactTemplate = {
		version: '1.0.0',
		user: '',
		modules: {
			shell: {
				'// example_darwin': {
					source: './shell/darwin.zshrc',
					target: '~/.zshrc',
					strategy: 'symlink'
				}
			},
			editor: {
				'// example_neovim': {
					source: './editor/nvim/',
					target: '~/.config/nvim',
					strategy: 'symlink'
				},
				'// example_keybindings': {
					source: './editor/keybindings.json',
					target: '~/.config/Code/User/keybindings.json',
					strategy: 'symlink'
				},
				'// example_snippets': {
					source: './editor/snippets/',
					target: '~/.config/Code/User/snippets/',
					strategy: 'symlink'
				}
			},
			terminal: {},
			git: {
				'// example_config': {
					source: './git/.gitconfig',
					target: '~/.gitconfig',
					strategy: 'symlink'
				}
			},
			'cli-tools': {},
			scripts: {},
			dotfiles: {},
			ai: {
				prompts: {},
				agents: {},
				providers: {},
				mcp: {}
			},
			ricing: {
				themes: {},
				fonts: {},
				wallpapers: {},
				colors: {},
				icons: {}
			}
		},
		secrets: [
			'// ANTHROPIC_API_KEY',
			'// OPENAI_API_KEY'
		]
	};

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/');
			return;
		}

		// Check if repo already exists
		try {
			const github = new GitHubClient($auth.token!);
			const exists = await github.repoExists($auth.user!.login);
			
			if (exists) {
				// Returning user, go to dashboard
				goto('/dashboard');
				return;
			}
			
			// New user, show confirmation
			state = 'confirm';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to check repository status';
			state = 'error';
		}
	});

	async function createPact() {
		state = 'creating';
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;

			// Step 1: Create repository
			steps[0].status = 'active';
			steps = steps;
			
			await github.createRepo();
			
			// Wait for GitHub to initialize the repo
			await new Promise(resolve => setTimeout(resolve, 2000));
			
			steps[0].status = 'complete';
			steps = steps;

			// Step 2: Create pact.json
			steps[1].status = 'active';
			steps = steps;

			const config = { ...pactTemplate, user: username };
			await github.updateFile(
				username,
				'pact.json',
				JSON.stringify(config, null, 2),
				'Initialize pact.json'
			);

			steps[1].status = 'complete';
			steps = steps;

			state = 'complete';

			// Auto-redirect after a moment
			setTimeout(() => {
				goto('/dashboard');
			}, 1500);

		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create pact';
			state = 'error';
			
			// Mark current step as error
			const activeStep = steps.find(s => s.status === 'active');
			if (activeStep) {
				activeStep.status = 'error';
				steps = steps;
			}
		}
	}
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono flex items-center justify-center">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<div class="relative z-10 w-full max-w-lg mx-4">
		<!-- Checking state -->
		{#if state === 'checking'}
			<div class="text-center">
				<Loader2 size={32} class="animate-spin mx-auto text-emerald-400" />
				<p class="mt-4 text-zinc-400">Checking your account...</p>
			</div>
		{/if}

		<!-- Confirmation state -->
		{#if state === 'confirm'}
			<div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-8 space-y-8">
				<!-- Header -->
				<div class="text-center space-y-4">
					{#if $auth.user}
						<img
							src={$auth.user.avatar_url}
							alt={$auth.user.login}
							class="w-16 h-16 rounded-full mx-auto border-2 border-zinc-700"
						/>
						<div>
							<h1 class="text-xl font-bold">Welcome, {$auth.user.name || $auth.user.login}</h1>
							<p class="text-sm text-zinc-500">@{$auth.user.login}</p>
						</div>
					{/if}
				</div>

				<!-- What is Pact -->
				<div class="space-y-4">
					<h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider">What is Pact?</h2>
					<div class="space-y-3">
						<div class="flex items-start gap-3">
							<div class="w-8 h-8 bg-zinc-800 rounded-lg flex items-center justify-center flex-shrink-0">
								<Github size={16} class="text-zinc-400" />
							</div>
							<div>
								<p class="text-sm text-zinc-300">Your dev environment in a GitHub repo</p>
								<p class="text-xs text-zinc-500">One <code class="text-zinc-400">pact.json</code> manifest + your config files</p>
							</div>
						</div>
						<div class="flex items-start gap-3">
							<div class="w-8 h-8 bg-zinc-800 rounded-lg flex items-center justify-center flex-shrink-0">
								<RefreshCw size={16} class="text-zinc-400" />
							</div>
							<div>
								<p class="text-sm text-zinc-300">Sync to any machine</p>
								<p class="text-xs text-zinc-500">Run <code class="text-zinc-400">pact sync</code> to apply your configs anywhere</p>
							</div>
						</div>
						<div class="flex items-start gap-3">
							<div class="w-8 h-8 bg-zinc-800 rounded-lg flex items-center justify-center flex-shrink-0">
								<Key size={16} class="text-zinc-400" />
							</div>
							<div>
								<p class="text-sm text-zinc-300">Secrets stay local</p>
								<p class="text-xs text-zinc-500">API keys stored in your OS keychain, never in the repo</p>
							</div>
						</div>
					</div>
				</div>

				<!-- What we'll create -->
				<div class="space-y-3">
					<h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider">What we'll create</h2>
					<div class="bg-zinc-800/50 rounded-lg p-4 space-y-2">
						<div class="flex items-center gap-2">
							<Terminal size={14} class="text-emerald-400" />
							<code class="text-sm text-zinc-300">{$auth.user?.login}/my-pact</code>
						</div>
						<p class="text-xs text-zinc-500 pl-6">A GitHub repository with a starter pact.json template</p>
					</div>
				</div>

				<!-- Confirm button -->
				<div class="space-y-3">
					<button
						on:click={createPact}
						class="w-full py-3 bg-emerald-500 text-zinc-950 font-medium rounded-lg hover:bg-emerald-400 transition-colors"
					>
						Create my-pact repo
					</button>
					<p class="text-xs text-zinc-500 text-center">
						Public by default. You can change visibility anytime in GitHub settings.
					</p>
				</div>
			</div>
		{/if}

		<!-- Creating state -->
		{#if state === 'creating' || state === 'complete'}
			<div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-8 space-y-8">
				<div class="text-center">
					<h1 class="text-xl font-bold">
						{state === 'complete' ? 'Your pact is ready!' : 'Setting up your pact...'}
					</h1>
				</div>

				<!-- Progress steps -->
				<div class="space-y-4">
					{#each steps as step}
						<div class="flex items-center gap-3">
							<div class="w-6 h-6 flex items-center justify-center">
								{#if step.status === 'complete'}
									<div class="w-6 h-6 bg-emerald-500 rounded-full flex items-center justify-center">
										<Check size={14} class="text-zinc-950" />
									</div>
								{:else if step.status === 'active'}
									<Loader2 size={20} class="animate-spin text-emerald-400" />
								{:else if step.status === 'error'}
									<div class="w-6 h-6 bg-red-500 rounded-full flex items-center justify-center">
										<span class="text-xs text-white">!</span>
									</div>
								{:else}
									<div class="w-6 h-6 border-2 border-zinc-700 rounded-full"></div>
								{/if}
							</div>
							<span class="text-sm {step.status === 'complete' ? 'text-zinc-300' : step.status === 'active' ? 'text-emerald-400' : 'text-zinc-500'}">
								{step.label}
							</span>
						</div>
					{/each}
				</div>

				{#if state === 'complete'}
					<div class="text-center">
						<p class="text-sm text-zinc-400">Redirecting to dashboard...</p>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Error state -->
		{#if state === 'error'}
			<div class="bg-zinc-900/50 border border-red-900/50 rounded-2xl p-8 space-y-6">
				<div class="text-center space-y-2">
					<div class="w-12 h-12 bg-red-500/20 rounded-full flex items-center justify-center mx-auto">
						<span class="text-red-400 text-xl">!</span>
					</div>
					<h1 class="text-xl font-bold text-red-400">Something went wrong</h1>
					<p class="text-sm text-zinc-400">{error}</p>
				</div>

				<div class="flex gap-3">
					<button
						on:click={() => { state = 'confirm'; error = ''; steps = steps.map(s => ({ ...s, status: 'pending' })); }}
						class="flex-1 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
					>
						Try again
					</button>
					<button
						on:click={() => goto('/')}
						class="flex-1 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
					>
						Back to home
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
