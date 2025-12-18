<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores/auth';
	import { getAuthUrl } from '$lib/github';
	import { FolderGit2, Github, Terminal, Cpu, Palette, Key } from 'lucide-svelte';

	onMount(() => {
		if ($isAuthenticated) {
			goto('/dashboard');
		}
	});

	function handleLogin() {
		window.location.href = getAuthUrl();
	}

	const features = [
		{ icon: Terminal, title: 'Shell configs', description: 'zsh, bash, PowerShell' },
		{ icon: Cpu, title: 'Editor setup', description: 'nvim, vscode, cursor' },
		{ icon: Palette, title: 'Themes', description: 'Terminal & editor themes' },
		{ icon: Key, title: 'Secrets', description: 'Safe in your OS keychain' }
	];
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<div class="relative z-10">
		<!-- Header -->
		<header class="p-6 flex justify-between items-center max-w-6xl mx-auto">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 flex items-center justify-center">
					<img src="/pact-clear-logo.png" alt="Pact" class="w-10 h-10" />
				</div>
				<span class="text-xl font-bold">pact</span>
			</div>
		</header>

		<!-- Hero -->
		<main class="max-w-6xl mx-auto px-6 py-24">
			<div class="text-center space-y-8">
				<div class="space-y-4">
					<h1 class="text-5xl md:text-6xl font-bold tracking-tight">
						Your portable
						<span
							class="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-emerald-600"
						>
							dev identity
						</span>
					</h1>
					<p class="text-xl text-zinc-400 max-w-2xl mx-auto">
						Shell, editor, AI prefs, themes — one kit, any machine. Edit in the browser, sync from
						terminal.
					</p>
				</div>

				<div class="flex justify-center gap-4">
					<button
						on:click={handleLogin}
						class="flex items-center gap-2 px-6 py-3 bg-white text-zinc-900 rounded-lg font-medium hover:bg-zinc-100 transition-all"
					>
						<Github size={20} />
						<span>Sign in with GitHub</span>
					</button>
				</div>

				<p class="text-sm text-zinc-500">
					New here? Signing in will create your pact repo automatically.
				</p>
			</div>

			<!-- Features grid -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-24">
				{#each features as feature}
					<div class="p-6 bg-zinc-900/30 border border-zinc-800/50 rounded-xl">
						<svelte:component this={feature.icon} size={24} class="text-emerald-400 mb-4" />
						<h3 class="font-medium mb-1">{feature.title}</h3>
						<p class="text-sm text-zinc-500">{feature.description}</p>
					</div>
				{/each}
			</div>

			<!-- How it works -->
			<div class="mt-24 space-y-8">
				<h2 class="text-2xl font-bold text-center">How it works</h2>
				<div class="grid md:grid-cols-3 gap-8">
					<div class="text-center space-y-4">
						<div
							class="w-12 h-12 bg-zinc-800 rounded-full flex items-center justify-center mx-auto text-emerald-400 font-bold"
						>
							1
						</div>
						<h3 class="font-medium">Connect GitHub</h3>
						<p class="text-sm text-zinc-500">
							We create a <code class="text-zinc-400">username/my-pact</code> repo to store your configs
						</p>
					</div>
					<div class="text-center space-y-4">
						<div
							class="w-12 h-12 bg-zinc-800 rounded-full flex items-center justify-center mx-auto text-emerald-400 font-bold"
						>
							2
						</div>
						<h3 class="font-medium">Configure</h3>
						<p class="text-sm text-zinc-500">
							Add your shell, editor, and tool configs through the web UI
						</p>
					</div>
					<div class="text-center space-y-4">
						<div
							class="w-12 h-12 bg-zinc-800 rounded-full flex items-center justify-center mx-auto text-emerald-400 font-bold"
						>
							3
						</div>
						<h3 class="font-medium">Sync anywhere</h3>
						<p class="text-sm text-zinc-500">
							Run <code class="text-zinc-400">pact sync</code> on any machine to apply your setup
						</p>
					</div>
				</div>
			</div>

			<!-- CLI preview -->
			<div class="mt-24">
				<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6 font-mono text-sm max-w-2xl mx-auto">
					<div class="flex items-center gap-2 mb-4">
						<div class="w-3 h-3 rounded-full bg-red-500"></div>
						<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<span class="ml-2 text-zinc-500">terminal</span>
					</div>
					<div class="space-y-2">
						<p><span class="text-emerald-400">$</span> pact init</p>
						<p class="text-zinc-500">Authenticating with GitHub...</p>
						<p class="text-zinc-500">Cloning to ~/.pact/...</p>
						<p class="text-emerald-400">✓ Pact initialized!</p>
						<p class="mt-4"><span class="text-emerald-400">$</span> pact sync</p>
						<p class="text-zinc-500">Syncing all modules...</p>
						<p class="text-emerald-400">✓ 5 synced, 0 failed</p>
					</div>
				</div>
			</div>
		</main>

		<!-- Footer -->
		<footer class="p-6 text-center text-sm text-zinc-500">
			<a href="https://github.com/cloudboy-jh/pact" class="hover:text-zinc-300 transition-colors">
				GitHub
			</a>
		</footer>
	</div>
</div>
