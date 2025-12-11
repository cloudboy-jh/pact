<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { GitHubClient, type PactConfig } from '$lib/github';
	import {
		Settings,
		Terminal,
		Key,
		Check,
		AlertCircle,
		ChevronRight,
		Cpu,
		Palette,
		MessageSquare,
		FolderGit2,
		RefreshCw,
		LogOut,
		Github
	} from 'lucide-svelte';

	let pactConfig: PactConfig | null = null;
	let loading = true;
	let error = '';

	const moduleIcons: Record<string, typeof Terminal> = {
		shell: Terminal,
		editor: Cpu,
		terminal: Terminal,
		git: Github,
		ai: MessageSquare,
		tools: Cpu,
		keybindings: Cpu,
		snippets: Cpu,
		fonts: Palette
	};

	const moduleDescriptions: Record<string, string> = {
		shell: 'Shell configuration',
		editor: 'Editor settings',
		terminal: 'Terminal emulator',
		git: 'Git configuration',
		ai: 'AI providers & prompts',
		tools: 'CLI tool configs',
		keybindings: 'Keyboard shortcuts',
		snippets: 'Code snippets',
		fonts: 'Font preferences'
	};

	let initialized = false;

	onMount(async () => {
		if (initialized) return; // Prevent reinit
		initialized = true;

		await auth.initialize();
		
		if (!$isAuthenticated) {
			goto('/');
			return;
		}

		await loadPactConfig();
	});

	async function loadPactConfig() {
		loading = true;
		error = '';
		
		try {
			const github = new GitHubClient($auth.token!);
			pactConfig = await github.getPactConfig($auth.user!.login);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load configuration';
		} finally {
			loading = false;
		}
	}

	function getModuleStatus(moduleName: string): string {
		if (!pactConfig) return 'not_configured';
		const module = (pactConfig.modules as Record<string, unknown>)[moduleName];
		if (!module || (typeof module === 'object' && Object.keys(module as object).length === 0)) {
			return 'not_configured';
		}
		return 'synced';
	}

	function getModuleFileCount(moduleName: string): number {
		// This would ideally come from the GitHub API
		return 0;
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'synced':
				return 'text-emerald-400';
			case 'pending':
				return 'text-amber-400';
			case 'missing':
				return 'text-amber-400';
			case 'not_configured':
				return 'text-zinc-500';
			default:
				return 'text-zinc-500';
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'synced':
				return Check;
			case 'pending':
				return AlertCircle;
			case 'missing':
				return AlertCircle;
			default:
				return null;
		}
	}

	function getStatusText(status: string): string {
		switch (status) {
			case 'synced':
				return 'synced';
			case 'pending':
				return 'pending';
			case 'missing':
				return 'missing';
			case 'not_configured':
				return 'not configured';
			default:
				return status;
		}
	}

	function handleLogout() {
		auth.logout();
		goto('/');
	}

	function navigateToEditor(moduleId: string) {
		goto(`/editor/${moduleId}`);
	}

	const modules = ['shell', 'editor', 'terminal', 'git', 'ai', 'tools', 'keybindings', 'snippets', 'fonts'];
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<div class="relative z-10 max-w-6xl mx-auto p-8">
		<!-- Header -->
		<header class="flex items-center justify-between mb-12">
			<div class="flex items-center gap-4">
				<div class="w-10 h-10 flex items-center justify-center">
					<img src="/pact-clear-logo.png" alt="Pact" class="w-10 h-10" />
				</div>
				<div>
					<h1 class="text-xl font-bold tracking-tight">pact</h1>
					{#if $auth.user}
						<p class="text-xs text-zinc-500">{$auth.user.login}/my-pact</p>
					{/if}
				</div>
			</div>

			<div class="flex items-center gap-3">
				<button
					on:click={loadPactConfig}
					class="flex items-center gap-2 px-4 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-sm hover:bg-zinc-800 hover:border-zinc-700 transition-all"
				>
					<RefreshCw size={14} class={loading ? 'animate-spin' : ''} />
					<span>Refresh</span>
				</button>
				<button
					on:click={handleLogout}
					class="flex items-center gap-2 p-2 bg-zinc-900 border border-zinc-800 rounded-lg hover:bg-zinc-800 transition-all"
					title="Logout"
				>
					<LogOut size={16} />
				</button>
			</div>
		</header>

		{#if loading}
			<div class="flex items-center justify-center py-24">
				<div
					class="animate-spin w-8 h-8 border-2 border-emerald-400 border-t-transparent rounded-full"
				></div>
			</div>
		{:else if error}
			<div class="text-center py-24">
				<p class="text-red-400">{error}</p>
				<button
					on:click={loadPactConfig}
					class="mt-4 text-emerald-400 hover:text-emerald-300 transition-colors"
				>
					Try again
				</button>
			</div>
		{:else}
			<!-- Status bar -->
			<div class="flex items-center gap-6 mb-8 p-4 bg-zinc-900/50 border border-zinc-800/50 rounded-xl">
				<div class="flex items-center gap-2">
					<div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
					<span class="text-sm text-zinc-400">Connected</span>
				</div>
				<div class="h-4 w-px bg-zinc-800"></div>
				<span class="text-sm text-zinc-500">
					Version: <span class="text-zinc-300">{pactConfig?.version || '1.0.0'}</span>
				</span>
			</div>

			<div class="grid grid-cols-3 gap-6">
				<!-- Kit Modules -->
				<div class="col-span-2 space-y-4">
					<h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider mb-4">Your Kit</h2>

					<div class="space-y-2">
						{#each modules as moduleId}
							{@const status = getModuleStatus(moduleId)}
							{@const StatusIcon = getStatusIcon(status)}
							<button
								on:click={() => navigateToEditor(moduleId)}
								class="w-full group flex items-center justify-between p-4 bg-zinc-900/30 border border-zinc-800/50 rounded-xl hover:bg-zinc-900/60 hover:border-zinc-700/50 transition-all text-left"
							>
								<div class="flex items-center gap-4">
									<div
										class="w-10 h-10 bg-zinc-800 rounded-lg flex items-center justify-center group-hover:bg-zinc-700 transition-colors"
									>
										<svelte:component this={moduleIcons[moduleId]} size={18} class="text-zinc-400" />
									</div>
									<div>
										<div class="flex items-center gap-2">
											<span class="font-medium capitalize">{moduleId}</span>
										</div>
										<span class="text-sm text-zinc-500">{moduleDescriptions[moduleId]}</span>
									</div>
								</div>

								<div class="flex items-center gap-3">
									<span class="flex items-center gap-1.5 text-xs {getStatusColor(status)}">
										{#if StatusIcon}
											<svelte:component this={StatusIcon} size={12} />
										{/if}
										{getStatusText(status)}
									</span>
									<ChevronRight
										size={16}
										class="text-zinc-600 group-hover:text-zinc-400 transition-colors"
									/>
								</div>
							</button>
						{/each}
					</div>
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Secrets -->
					<div>
						<h2
							class="text-sm font-medium text-zinc-400 uppercase tracking-wider mb-4 flex items-center gap-2"
						>
							<Key size={14} />
							Secrets
						</h2>

						<div class="bg-zinc-900/30 border border-zinc-800/50 rounded-xl overflow-hidden">
							{#if pactConfig?.secrets && pactConfig.secrets.length > 0}
								{#each pactConfig.secrets as secret, i}
									<div
										class="flex items-center justify-between p-3 {i !== pactConfig.secrets.length - 1
											? 'border-b border-zinc-800/50'
											: ''}"
									>
										<div class="flex items-center gap-2">
											<div class="w-1.5 h-1.5 bg-zinc-600 rounded-full"></div>
											<span class="text-xs font-mono text-zinc-300">{secret}</span>
										</div>
										<span class="text-xs text-zinc-500">local only</span>
									</div>
								{/each}
							{:else}
								<div class="p-3 text-xs text-zinc-500">No secrets configured</div>
							{/if}
						</div>
					</div>

					<!-- Quick Actions -->
					<div class="p-4 bg-zinc-900/30 border border-zinc-800/50 rounded-xl">
						<p class="text-xs text-zinc-500 font-mono mb-3">Quick commands</p>
						<div class="space-y-2 text-sm font-mono">
							<div class="text-zinc-400">$ pact sync</div>
							<div class="text-zinc-400">$ pact push</div>
							<div class="text-zinc-400">$ pact status</div>
						</div>
					</div>

					<!-- User Info -->
					{#if $auth.user}
						<div class="p-4 bg-zinc-900/30 border border-zinc-800/50 rounded-xl">
							<div class="flex items-center gap-3">
								<img
									src={$auth.user.avatar_url}
									alt={$auth.user.login}
									class="w-10 h-10 rounded-full"
								/>
								<div>
									<p class="text-sm font-medium">{$auth.user.name || $auth.user.login}</p>
									<p class="text-xs text-zinc-500">@{$auth.user.login}</p>
								</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>
