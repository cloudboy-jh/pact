<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { GitHubClient, type PactConfig } from '$lib/github';
	import {
		Terminal,
		Key,
		Check,
		AlertCircle,
		ChevronRight,
		ChevronDown,
		Cpu,
		Palette,
		MessageSquare,
		RefreshCw,
		LogOut,
		Github,
		Sparkles,
		Bot,
		Server,
		Paintbrush,
		Type,
		Image,
		Droplet,
		Shapes,
		FileCode,
		FolderDot,
		ScrollText
	} from 'lucide-svelte';

	let pactConfig: PactConfig | null = null;
	let loading = true;
	let error = '';

	// Dropdown state
	let aiExpanded = false;
	let ricingExpanded = false;

	// Top-level modules - support both old (cli-tools) and new (cli) formats
	const topLevelModules = [
		{ id: 'shell', icon: Terminal, description: 'Shell configuration', altIds: [] },
		{ id: 'editor', icon: Cpu, description: 'Editor settings, keybindings, snippets', altIds: [] },
		{ id: 'terminal', icon: Terminal, description: 'Terminal emulator', altIds: [] },
		{ id: 'git', icon: Github, description: 'Git configuration', altIds: [] },
		{ id: 'cli', icon: FileCode, description: 'CLI tool configs', altIds: ['cli-tools'] },
		{ id: 'scripts', icon: ScrollText, description: 'Personal utility scripts', altIds: [] },
		{ id: 'dotfiles', icon: FolderDot, description: 'Misc dotfiles', altIds: [] },
		{ id: 'apps', icon: Cpu, description: 'Application shortcuts', altIds: [] }
	];

	// AI/LLM dropdown sub-items - support both ai.* and llm.* formats
	const aiSubItems = [
		{ id: 'llm.providers', label: 'Providers', icon: Key, description: 'API provider configs', altIds: ['ai.providers'] },
		{ id: 'llm.coding', label: 'Coding', icon: Bot, description: 'Coding agents & models', altIds: ['ai.agents'] },
		{ id: 'llm.chat', label: 'Chat', icon: MessageSquare, description: 'Chat providers', altIds: ['ai.prompts'] },
		{ id: 'llm.local', label: 'Local', icon: Server, description: 'Local models (Ollama)', altIds: ['ai.mcp'] }
	];

	// Ricing dropdown sub-items
	const ricingSubItems = [
		{ id: 'ricing.themes', label: 'Themes', icon: Paintbrush, description: 'Terminal, editor themes', altIds: [] },
		{ id: 'ricing.fonts', label: 'Fonts', icon: Type, description: 'Font preferences', altIds: [] },
		{ id: 'ricing.wallpapers', label: 'Wallpapers/PFPs', icon: Image, description: 'Backgrounds, profile pics', altIds: [] },
		{ id: 'ricing.colors', label: 'Colors', icon: Droplet, description: 'Color palettes', altIds: [] },
		{ id: 'ricing.icons', label: 'Icons', icon: Shapes, description: 'Icon packs', altIds: [] }
	];

	let initialized = false;

	onMount(async () => {
		if (initialized) return;
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

	function checkPath(config: Record<string, unknown>, path: string): unknown {
		const parts = path.split('.');
		let current: unknown = config;
		
		for (const part of parts) {
			if (current && typeof current === 'object' && part in current) {
				current = (current as Record<string, unknown>)[part];
			} else {
				return null;
			}
		}
		return current;
	}

	function getModuleStatus(modulePath: string, altIds: string[] = []): string {
		if (!pactConfig) return 'not_configured';
		
		// Try all possible paths: primary id, alt ids, and modules.* versions
		const pathsToTry = [
			modulePath,
			...altIds,
			`modules.${modulePath}`,
			...altIds.map(alt => `modules.${alt}`)
		];
		
		for (const path of pathsToTry) {
			const result = checkPath(pactConfig as Record<string, unknown>, path);
			if (result !== null) {
				// Found something - check if it has valid content
				if (typeof result === 'object' && result !== null) {
					const keys = Object.keys(result).filter(k => !k.startsWith('//'));
					if (keys.length > 0) {
						return 'synced';
					}
				} else if (result !== undefined) {
					// Non-object value (string, array, etc.) counts as configured
					return 'synced';
				}
			}
		}
		
		return 'not_configured';
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'synced':
				return 'text-emerald-400';
			case 'pending':
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

	function navigateToEditor(sectionId: string) {
		goto(`/editor?section=${sectionId}`);
	}

	function toggleAi() {
		aiExpanded = !aiExpanded;
	}

	function toggleRicing() {
		ricingExpanded = !ricingExpanded;
	}
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
						<!-- Top-level modules -->
						{#each topLevelModules as module}
							{@const status = getModuleStatus(module.id, module.altIds)}
							{@const StatusIcon = getStatusIcon(status)}
							<button
								on:click={() => navigateToEditor(module.id)}
								class="w-full group flex items-center justify-between p-4 bg-zinc-900/30 border border-zinc-800/50 rounded-xl hover:bg-zinc-900/60 hover:border-zinc-700/50 transition-all text-left"
							>
								<div class="flex items-center gap-4">
									<div
										class="w-10 h-10 bg-zinc-800 rounded-lg flex items-center justify-center group-hover:bg-zinc-700 transition-colors"
									>
										<svelte:component this={module.icon} size={18} class="text-zinc-400" />
									</div>
									<div>
										<span class="font-medium capitalize">{module.id.replace('-', ' ')}</span>
										<p class="text-sm text-zinc-500">{module.description}</p>
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

						<!-- AI Dropdown -->
						<div class="rounded-xl border border-zinc-800/50 overflow-hidden">
							<button
								on:click={toggleAi}
								class="w-full group flex items-center justify-between p-4 bg-zinc-900/30 hover:bg-zinc-900/60 transition-all text-left"
							>
								<div class="flex items-center gap-4">
									<div
										class="w-10 h-10 bg-zinc-800 rounded-lg flex items-center justify-center group-hover:bg-zinc-700 transition-colors"
									>
										<Sparkles size={18} class="text-zinc-400" />
									</div>
									<div>
										<span class="font-medium">AI</span>
										<p class="text-sm text-zinc-500">Prompts, agents, providers</p>
									</div>
								</div>

								<div class="flex items-center gap-3">
									{#if aiExpanded}
										<ChevronDown size={16} class="text-zinc-500" />
									{:else}
										<ChevronRight size={16} class="text-zinc-500" />
									{/if}
								</div>
							</button>

							{#if aiExpanded}
								<div class="border-t border-zinc-800/50 bg-zinc-950/50">
									{#each aiSubItems as item}
										{@const status = getModuleStatus(item.id, item.altIds)}
										{@const StatusIcon = getStatusIcon(status)}
										<button
											on:click={() => navigateToEditor(item.id)}
											class="w-full group flex items-center justify-between p-3 pl-8 hover:bg-zinc-900/60 transition-all text-left border-b border-zinc-800/30 last:border-b-0"
										>
											<div class="flex items-center gap-3">
												<div
													class="w-8 h-8 bg-zinc-800/50 rounded-lg flex items-center justify-center group-hover:bg-zinc-700/50 transition-colors"
												>
													<svelte:component this={item.icon} size={14} class="text-zinc-500" />
												</div>
												<div>
													<span class="text-sm font-medium text-zinc-300">{item.label}</span>
													<p class="text-xs text-zinc-600">{item.description}</p>
												</div>
											</div>

											<div class="flex items-center gap-3">
												<span class="flex items-center gap-1.5 text-xs {getStatusColor(status)}">
													{#if StatusIcon}
														<svelte:component this={StatusIcon} size={10} />
													{/if}
													{getStatusText(status)}
												</span>
												<ChevronRight
													size={14}
													class="text-zinc-700 group-hover:text-zinc-500 transition-colors"
												/>
											</div>
										</button>
									{/each}
								</div>
							{/if}
						</div>

						<!-- Ricing Dropdown -->
						<div class="rounded-xl border border-zinc-800/50 overflow-hidden">
							<button
								on:click={toggleRicing}
								class="w-full group flex items-center justify-between p-4 bg-zinc-900/30 hover:bg-zinc-900/60 transition-all text-left"
							>
								<div class="flex items-center gap-4">
									<div
										class="w-10 h-10 bg-zinc-800 rounded-lg flex items-center justify-center group-hover:bg-zinc-700 transition-colors"
									>
										<Palette size={18} class="text-zinc-400" />
									</div>
									<div>
										<span class="font-medium">Ricing</span>
										<p class="text-sm text-zinc-500">Themes, fonts, wallpapers</p>
									</div>
								</div>

								<div class="flex items-center gap-3">
									{#if ricingExpanded}
										<ChevronDown size={16} class="text-zinc-500" />
									{:else}
										<ChevronRight size={16} class="text-zinc-500" />
									{/if}
								</div>
							</button>

							{#if ricingExpanded}
								<div class="border-t border-zinc-800/50 bg-zinc-950/50">
									{#each ricingSubItems as item}
										{@const status = getModuleStatus(item.id, item.altIds)}
										{@const StatusIcon = getStatusIcon(status)}
										<button
											on:click={() => navigateToEditor(item.id)}
											class="w-full group flex items-center justify-between p-3 pl-8 hover:bg-zinc-900/60 transition-all text-left border-b border-zinc-800/30 last:border-b-0"
										>
											<div class="flex items-center gap-3">
												<div
													class="w-8 h-8 bg-zinc-800/50 rounded-lg flex items-center justify-center group-hover:bg-zinc-700/50 transition-colors"
												>
													<svelte:component this={item.icon} size={14} class="text-zinc-500" />
												</div>
												<div>
													<span class="text-sm font-medium text-zinc-300">{item.label}</span>
													<p class="text-xs text-zinc-600">{item.description}</p>
												</div>
											</div>

											<div class="flex items-center gap-3">
												<span class="flex items-center gap-1.5 text-xs {getStatusColor(status)}">
													{#if StatusIcon}
														<svelte:component this={StatusIcon} size={10} />
													{/if}
													{getStatusText(status)}
												</span>
												<ChevronRight
													size={14}
													class="text-zinc-700 group-hover:text-zinc-500 transition-colors"
												/>
											</div>
										</button>
									{/each}
								</div>
							{/if}
						</div>
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
