<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { GitHubClient, type GitHubFile } from '$lib/github';
	import {
		ArrowLeft,
		File,
		Folder,
		Plus,
		Save,
		Trash2,
		FolderGit2,
		X
	} from 'lucide-svelte';

	let moduleId = '';
	let files: GitHubFile[] = [];
	let selectedFile: GitHubFile | null = null;
	let fileContent = '';
	let originalContent = '';
	let loading = true;
	let saving = false;
	let error = '';
	let currentPath = '';

	// New file dialog state
	let showNewFileDialog = false;
	let newFileName = '';
	let newFileNameInput: HTMLInputElement;

	// OS tabs for modules that support OS-specific configs
	const osModules = ['shell', 'editor', 'terminal'];
	let selectedOS: 'darwin' | 'linux' | 'windows' = 'darwin';

	onMount(async () => {
		await auth.initialize();

		if (!$isAuthenticated) {
			goto('/');
			return;
		}

		moduleId = $page.params.module;
		await loadFiles();
	});

	async function loadFiles(path = '') {
		loading = true;
		error = '';
		// Always use moduleId as the base path
		currentPath = path || moduleId;

		try {
			const github = new GitHubClient($auth.token!);
			
			try {
				files = await github.getContents($auth.user!.login, currentPath);
			} catch {
				// Module directory doesn't exist yet
				files = [];
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load files';
		} finally {
			loading = false;
		}
	}

	async function selectFile(file: GitHubFile) {
		if (file.type === 'dir') {
			await loadFiles(file.path);
			return;
		}

		loading = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			fileContent = await github.getFileContent($auth.user!.login, file.path);
			originalContent = fileContent;
			selectedFile = file;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load file';
		} finally {
			loading = false;
		}
	}

	async function saveFile() {
		if (!selectedFile) return;

		saving = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const result = await github.updateFile(
				$auth.user!.login,
				selectedFile.path,
				fileContent,
				`Update ${selectedFile.name}`,
				selectedFile.sha
			);
			
			originalContent = fileContent;
			
			// Update selectedFile with new SHA from response
			selectedFile = {
				...selectedFile,
				sha: result.content.sha
			};
			
			// Refresh file list
			await loadFiles(currentPath);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save file';
		} finally {
			saving = false;
		}
	}

	function openNewFileDialog() {
		newFileName = '';
		showNewFileDialog = true;
		// Focus the input after the dialog renders
		setTimeout(() => newFileNameInput?.focus(), 0);
	}

	function closeNewFileDialog() {
		showNewFileDialog = false;
		newFileName = '';
	}

	async function createFile() {
		const filename = newFileName.trim();
		if (!filename) return;

		closeNewFileDialog();
		saving = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const path = `${currentPath}/${filename}`;
			
			await github.updateFile(
				$auth.user!.login,
				path,
				'',
				`Create ${filename}`
			);
			
			await loadFiles(currentPath);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create file';
		} finally {
			saving = false;
		}
	}

	function handleNewFileKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && newFileName.trim()) {
			createFile();
		} else if (event.key === 'Escape') {
			closeNewFileDialog();
		}
	}

	async function deleteFile() {
		if (!selectedFile) return;
		
		if (!confirm(`Delete ${selectedFile.name}?`)) return;

		saving = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			await github.deleteFile(
				$auth.user!.login,
				selectedFile.path,
				selectedFile.sha,
				`Delete ${selectedFile.name}`
			);
			
			selectedFile = null;
			fileContent = '';
			originalContent = '';
			
			await loadFiles(currentPath);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete file';
		} finally {
			saving = false;
		}
	}

	function goBack() {
		if (currentPath !== moduleId) {
			const parentPath = currentPath.split('/').slice(0, -1).join('/');
			loadFiles(parentPath || moduleId);
			selectedFile = null;
			fileContent = '';
			originalContent = '';
		}
	}

	function goToDashboard() {
		goto('/dashboard');
	}

	$: hasChanges = fileContent !== originalContent;
	$: isOSModule = osModules.includes(moduleId);
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<!-- New File Dialog -->
	{#if showNewFileDialog}
		<div class="fixed inset-0 z-50 flex items-center justify-center">
			<!-- Backdrop -->
			<button
				class="absolute inset-0 bg-black/60 backdrop-blur-sm"
				on:click={closeNewFileDialog}
				aria-label="Close dialog"
			></button>
			
			<!-- Dialog -->
			<div class="relative bg-zinc-900 border border-zinc-700 rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
				<!-- Header -->
				<div class="flex items-center justify-between p-4 border-b border-zinc-800">
					<h2 class="text-lg font-semibold">Create New File</h2>
					<button
						on:click={closeNewFileDialog}
						class="p-1 hover:bg-zinc-800 rounded-lg transition-colors text-zinc-400 hover:text-zinc-200"
					>
						<X size={18} />
					</button>
				</div>
				
				<!-- Content -->
				<div class="p-4 space-y-4">
					<div>
						<label for="filename" class="block text-sm text-zinc-400 mb-2">Filename</label>
						<input
							bind:this={newFileNameInput}
							bind:value={newFileName}
							on:keydown={handleNewFileKeydown}
							type="text"
							id="filename"
							placeholder="example.txt"
							class="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 placeholder-zinc-500"
						/>
					</div>
					<p class="text-xs text-zinc-500">
						File will be created in: <code class="text-zinc-400">{currentPath}/</code>
					</p>
				</div>
				
				<!-- Footer -->
				<div class="flex justify-end gap-2 p-4 border-t border-zinc-800 bg-zinc-900/50">
					<button
						on:click={closeNewFileDialog}
						class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 rounded-lg transition-colors"
					>
						Cancel
					</button>
					<button
						on:click={createFile}
						disabled={!newFileName.trim()}
						class="px-4 py-2 text-sm bg-emerald-500 text-zinc-950 font-medium rounded-lg hover:bg-emerald-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Create File
					</button>
				</div>
			</div>
		</div>
	{/if}

	<div class="relative z-10 h-screen flex flex-col">
		<!-- Header -->
		<header class="flex items-center justify-between p-4 border-b border-zinc-800">
			<div class="flex items-center gap-4">
				<button
					on:click={goToDashboard}
					class="p-2 hover:bg-zinc-800 rounded-lg transition-colors"
				>
					<ArrowLeft size={20} />
				</button>
				<div class="flex items-center gap-3">
					<div
						class="w-8 h-8 bg-gradient-to-br from-emerald-400 to-emerald-600 rounded-lg flex items-center justify-center"
					>
						<FolderGit2 size={16} class="text-zinc-950" />
					</div>
					<div>
						<h1 class="font-bold capitalize">{moduleId}</h1>
						<p class="text-xs text-zinc-500">{currentPath}</p>
					</div>
				</div>
			</div>

			<div class="flex items-center gap-2">
				{#if selectedFile}
					<button
						on:click={deleteFile}
						disabled={saving}
						class="flex items-center gap-2 px-3 py-2 text-red-400 hover:bg-red-400/10 rounded-lg transition-colors text-sm disabled:opacity-50"
					>
						<Trash2 size={14} />
						Delete
					</button>
					<button
						on:click={saveFile}
						disabled={saving || !hasChanges}
						class="flex items-center gap-2 px-4 py-2 bg-emerald-500 text-zinc-950 rounded-lg font-medium hover:bg-emerald-400 transition-colors text-sm disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Save size={14} />
						{saving ? 'Saving...' : 'Save'}
					</button>
				{/if}
			</div>
		</header>

		<!-- OS Tabs for supported modules -->
		{#if isOSModule}
			<div class="flex gap-1 p-2 border-b border-zinc-800 bg-zinc-900/50">
				{#each ['darwin', 'linux', 'windows'] as os}
					<button
						on:click={() => selectedOS = os}
						class="px-3 py-1.5 text-sm rounded-lg transition-colors {selectedOS === os
							? 'bg-zinc-800 text-white'
							: 'text-zinc-500 hover:text-zinc-300'}"
					>
						{os === 'darwin' ? 'macOS' : os === 'linux' ? 'Linux' : 'Windows'}
					</button>
				{/each}
			</div>
		{/if}

		<div class="flex-1 flex overflow-hidden">
			<!-- File list sidebar -->
			<div class="w-64 border-r border-zinc-800 flex flex-col">
				<div class="p-2 border-b border-zinc-800 flex items-center justify-between">
					{#if currentPath !== moduleId}
						<button
							on:click={goBack}
							class="flex items-center gap-1 text-sm text-zinc-400 hover:text-zinc-200 transition-colors"
						>
							<ArrowLeft size={14} />
							Back
						</button>
					{:else}
						<span class="text-sm text-zinc-500">Files</span>
					{/if}
					<button
						on:click={openNewFileDialog}
						class="p-1 hover:bg-zinc-800 rounded transition-colors"
						title="New file"
					>
						<Plus size={16} />
					</button>
				</div>

				<div class="flex-1 overflow-y-auto">
					{#if loading && !selectedFile}
						<div class="p-4 text-center text-zinc-500">Loading...</div>
					{:else if files.length === 0}
						<div class="p-4 text-center text-zinc-500 text-sm">
							No files yet. Click + to create one.
						</div>
					{:else}
						{#each files as file}
							<button
								on:click={() => selectFile(file)}
								class="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-zinc-800/50 transition-colors {selectedFile?.path ===
								file.path
									? 'bg-zinc-800 text-white'
									: 'text-zinc-400'}"
							>
								{#if file.type === 'dir'}
									<Folder size={14} />
								{:else}
									<File size={14} />
								{/if}
								<span class="truncate">{file.name}</span>
							</button>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Editor area -->
			<div class="flex-1 flex flex-col">
				{#if error}
					<div class="p-4 bg-red-500/10 border-b border-red-500/20 text-red-400 text-sm">
						{error}
					</div>
				{/if}

				{#if selectedFile}
					<div class="p-2 border-b border-zinc-800 bg-zinc-900/50 text-sm text-zinc-400">
						{selectedFile.path}
						{#if hasChanges}
							<span class="text-amber-400 ml-2">• unsaved</span>
						{/if}
					</div>
					<textarea
						bind:value={fileContent}
						class="flex-1 w-full bg-transparent p-4 resize-none focus:outline-none text-sm leading-relaxed"
						spellcheck="false"
						placeholder="Enter file content..."
					></textarea>
				{:else}
					<div class="flex-1 flex items-center justify-center text-zinc-500">
						Select a file to edit
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
