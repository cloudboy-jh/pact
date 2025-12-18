<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';

	onMount(async () => {
		await auth.initialize();
	});
</script>

<svelte:head>
	<title>Pact - Your portable dev identity</title>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet" />
</svelte:head>

{#if $auth.loading}
	<div class="min-h-screen bg-zinc-950 flex flex-col items-center justify-center">
		<img src="/pact-clear-logo.png" alt="Pact" class="w-16 h-16 mb-6" />
		<div class="animate-spin w-6 h-6 border-2 border-emerald-400 border-t-transparent rounded-full"></div>
	</div>
{:else}
	<slot />
{/if}
