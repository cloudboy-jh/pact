<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { GitHubClient, type GitHubFile } from '$lib/github';
	import CodeEditor from '$lib/components/CodeEditor.svelte';
import {
		ArrowLeft,
		File,
		Folder,
		FolderOpen,
		ChevronRight,
		ChevronDown,
		Plus,
		Save,
		Check,
		Loader2,
		X,
		FileJson,
		ExternalLink,
		Sparkles,
		Github
	} from 'lucide-svelte';

	// LLM providers for "Open in" dropdown
	const llmProviders = [
		{ id: 'claude', name: 'Claude', url: 'https://claude.ai/new', favicon: 'https://www.google.com/s2/favicons?domain=claude.ai&sz=32' },
		{ id: 'chatgpt', name: 'ChatGPT', url: 'https://chat.openai.com/', favicon: 'https://cdn.oaistatic.com/assets/favicon-o4x1jcxe.svg' },
		{ id: 'gemini', name: 'Gemini', url: 'https://gemini.google.com/app', favicon: 'https://www.google.com/s2/favicons?domain=gemini.google.com&sz=32' },
		{ id: 'grok', name: 'Grok', url: 'https://grok.x.ai/', favicon: 'https://www.google.com/s2/favicons?domain=x.ai&sz=32' }
	];

	let showLlmDropdown = false;
	let toastMessage = '';
	let showToast = false;

	// State
	let loading = true;
	let saving = false;
	let saveStatus: 'idle' | 'saving' | 'saved' | 'error' = 'idle';
	let error = '';

	// File tree
	let repoFiles: GitHubFile[] = [];
	let expandedFolders: Set<string> = new Set();

	// pact.json parsed sections - now supports nested structure
	interface PactSection {
		id: string;
		label: string;
		children?: PactSection[];
	}
	let pactSections: PactSection[] = [];
	let pactJsonExpanded = true;
	let expandedPactSections: Set<string> = new Set();

	// Current file being edited
	let currentFile: 'pact.json' | string = 'pact.json';
	let currentContent = '';
	let currentSha = '';

	// Section highlighting
	let highlightLines: { from: number; to: number } | null = null;
	let highlightDismissHint = false;

	// New file dialog
	let showNewFileDialog = false;
	let newFileName = '';
	let newFileNameInput: HTMLInputElement;

	// Auto-save debounce
	let saveTimeout: ReturnType<typeof setTimeout> | null = null;

	// Editor component reference
	let editorComponent: CodeEditor;

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/');
			return;
		}

		await loadPactJson();
		await loadRepoFiles();

		// Check for section param in URL
		const section = $page.url.searchParams.get('section');
		if (section) {
			highlightSection(section);
		}

		// Check for file param in URL
		const file = $page.url.searchParams.get('file');
		if (file) {
			await openFile(file);
		}
	});

	async function loadPactJson() {
		loading = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;

			// Get pact.json content and SHA
			const file = await github.getContents(username, 'pact.json');
			const pactFile = Array.isArray(file) ? file[0] : file;
			currentSha = pactFile.sha;

			currentContent = await github.getFileContent(username, 'pact.json');
			currentFile = 'pact.json';

			// Parse sections from pact.json - handle nested structure
			try {
				const parsed = JSON.parse(currentContent);
				if (parsed.modules) {
					pactSections = Object.entries(parsed.modules).map(([key, value]) => {
						const section: PactSection = { id: key, label: key };
						// Check if this module has nested sub-sections (like ai.prompts)
						if (value && typeof value === 'object' && !Array.isArray(value)) {
							const subKeys = Object.keys(value as object).filter(k => !k.startsWith('//'));
							if (subKeys.length > 0 && typeof (value as Record<string, unknown>)[subKeys[0]] === 'object') {
								section.children = subKeys.map(subKey => ({
									id: `${key}.${subKey}`,
									label: subKey
								}));
							}
						}
						return section;
					});
				}
			} catch {
				// Invalid JSON, no sections
				pactSections = [];
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load pact.json';
		} finally {
			loading = false;
		}
	}

	async function loadRepoFiles() {
		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;

			const contents = await github.getContents(username, '');
			// Filter out pact.json since it's shown separately
			repoFiles = contents.filter(f => f.name !== 'pact.json' && f.name !== 'README.md' && f.name !== '.gitkeep');
		} catch {
			// Repo might be empty or have issues
			repoFiles = [];
		}
	}

	async function loadFolderContents(folderPath: string): Promise<GitHubFile[]> {
		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;
			return await github.getContents(username, folderPath);
		} catch {
			return [];
		}
	}

	function toggleFolder(folderPath: string) {
		if (expandedFolders.has(folderPath)) {
			expandedFolders.delete(folderPath);
		} else {
			expandedFolders.add(folderPath);
		}
		expandedFolders = expandedFolders;
	}

	async function openFile(filePath: string) {
		if (currentFile === filePath) return;

		loading = true;
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;

			// Get file content and SHA
			const file = await github.getContents(username, filePath);
			const fileData = Array.isArray(file) ? file[0] : file;
			currentSha = fileData.sha;

			currentContent = await github.getFileContent(username, filePath);
			currentFile = filePath;

			// Clear any highlights when switching files
			highlightLines = null;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load file';
		} finally {
			loading = false;
		}
	}

	async function openPactJson() {
		if (currentFile === 'pact.json') return;
		await loadPactJson();
	}

	function highlightSection(sectionName: string) {
		if (currentFile !== 'pact.json') {
			// First open pact.json
			loadPactJson().then(() => {
				findAndHighlightSection(sectionName);
			});
		} else {
			findAndHighlightSection(sectionName);
		}
	}

	function findAndHighlightSection(sectionPath: string) {
		// Support nested paths like "ai.prompts"
		const parts = sectionPath.split('.');
		const lines = currentContent.split('\n');
		
		let startLine = -1;
		let endLine = -1;
		let depth = 0;
		let currentPartIndex = 0;
		let searchingForPart = true;
		let foundAllParts = false;

		for (let i = 0; i < lines.length; i++) {
			const line = lines[i];
			
			if (searchingForPart && currentPartIndex < parts.length) {
				const partPattern = new RegExp(`^\\s*"${parts[currentPartIndex]}"\\s*:`);
				
				if (partPattern.test(line)) {
					currentPartIndex++;
					
					if (currentPartIndex === parts.length) {
						// Found the final part, start tracking
						startLine = i + 1; // 1-indexed
						searchingForPart = false;
						foundAllParts = true;
						depth = (line.match(/{/g) || []).length - (line.match(/}/g) || []).length;
						if (depth === 0 && line.includes('{')) depth = 1;
					}
					continue;
				}
			}

			if (foundAllParts && !searchingForPart) {
				depth += (line.match(/{/g) || []).length;
				depth -= (line.match(/}/g) || []).length;

				if (depth <= 0) {
					endLine = i + 1; // 1-indexed
					break;
				}
			}
		}

		if (startLine > 0 && endLine > 0) {
			highlightLines = { from: startLine, to: endLine };
			highlightDismissHint = true;
		}
	}

	function handleEditorClick() {
		if (highlightLines) {
			highlightLines = null;
			highlightDismissHint = false;
			editorComponent?.clearHighlight();
		}
	}

	function handleContentChange(event: CustomEvent<string>) {
		currentContent = event.detail;

		// Update sections if editing pact.json
		if (currentFile === 'pact.json') {
			try {
				const parsed = JSON.parse(currentContent);
				if (parsed.modules) {
					pactSections = Object.keys(parsed.modules).map(key => ({ id: key, label: key }));
				}
			} catch {
				// Invalid JSON during editing, keep old sections
			}
		}

		// Auto-save with debounce
		if (saveTimeout) {
			clearTimeout(saveTimeout);
		}
		saveTimeout = setTimeout(() => {
			saveFile();
		}, 1500);
	}

	async function saveFile() {
		if (saving) return;

		saving = true;
		saveStatus = 'saving';
		error = '';

		try {
			const github = new GitHubClient($auth.token!);
			const username = $auth.user!.login;

			const result = await github.updateFile(
				username,
				currentFile,
				currentContent,
				`Update ${currentFile}`,
				currentSha
			);

			// Update SHA from response
			currentSha = result.content.sha;

			saveStatus = 'saved';

			// Reset status after a moment
			setTimeout(() => {
				if (saveStatus === 'saved') {
					saveStatus = 'idle';
				}
			}, 2000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save file';
			saveStatus = 'error';
		} finally {
			saving = false;
		}
	}

	// New file dialog
	function openNewFileDialog() {
		newFileName = '';
		showNewFileDialog = true;
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
			const username = $auth.user!.login;

			await github.updateFile(
				username,
				filename,
				'',
				`Create ${filename}`
			);

			await loadRepoFiles();

			// Open the new file
			await openFile(filename);
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

	function goToDashboard() {
		goto('/dashboard');
	}

	function generateLlmPrompt(): string {
		const username = $auth.user?.login || 'user';
		return `I'm using Pact (https://github.com/cloudboy-jh/pact) to manage my development environment configuration.

Pact stores dev environment configs (shell, editor, terminal, git, AI tools, themes, etc.) in a single pact.json manifest file in a GitHub repo called "my-pact".

Here's my current pact.json configuration:

\`\`\`json
${currentContent}
\`\`\`

Please help me edit this configuration. You can:
- Add new module configurations
- Modify existing settings
- Suggest improvements based on best practices
- Help me set up new tools or customize existing ones

The pact.json structure uses:
- "modules" for different config categories (shell, editor, terminal, git, cli-tools, scripts, dotfiles, ai, ricing)
- File references like "./shell/darwin.zshrc" that point to actual config files in the repo
- "source" for the file path in the repo, "target" for where it should be symlinked/copied on the system
- "strategy" can be "symlink" or "copy"

When suggesting changes, please provide the updated JSON that I can copy back into my pact.json file.`;
	}

	async function openInLlm(provider: typeof llmProviders[0]) {
		const prompt = generateLlmPrompt();
		
		// Open the URL first (must be synchronous with user action to avoid popup blocker)
		const newWindow = window.open(provider.url, '_blank');
		
		// Then copy prompt to clipboard
		try {
			await navigator.clipboard.writeText(prompt);
			toastMessage = `Prompt copied! Paste it in ${provider.name} (Ctrl+V / Cmd+V)`;
			showToast = true;
			setTimeout(() => { showToast = false; }, 4000);
		} catch {
			// Clipboard failed - show manual copy option
			toastMessage = `Opened ${provider.name} - copy prompt manually`;
			showToast = true;
			setTimeout(() => { showToast = false; }, 3000);
		}
		
		showLlmDropdown = false;
	}

	function toggleLlmDropdown() {
		showLlmDropdown = !showLlmDropdown;
		if (showLlmDropdown) {
			// Add click outside listener
			setTimeout(() => {
				window.addEventListener('click', closeLlmDropdown, { once: true });
			}, 0);
		}
	}

	function closeLlmDropdown() {
		showLlmDropdown = false;
	}

	$: isJson = currentFile.endsWith('.json');
</script>

<div class="h-screen bg-zinc-950 text-zinc-100 font-mono flex flex-col">
	<!-- Subtle grid background -->
	<div
		class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"
	></div>

	<!-- Toast Notification -->
	{#if showToast}
		<div class="fixed bottom-6 right-6 z-50 animate-in slide-in-from-bottom-2 fade-in duration-200">
			<div class="flex items-center gap-2 px-4 py-3 bg-emerald-500/90 text-zinc-950 rounded-lg shadow-lg font-medium text-sm">
				<Check size={16} />
				{toastMessage}
			</div>
		</div>
	{/if}

	<!-- New File Dialog -->
	{#if showNewFileDialog}
		<div class="fixed inset-0 z-50 flex items-center justify-center">
			<button
				class="absolute inset-0 bg-black/60 backdrop-blur-sm"
				on:click={closeNewFileDialog}
				aria-label="Close dialog"
			></button>
			
			<div class="relative bg-zinc-900 border border-zinc-700 rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
				<div class="flex items-center justify-between p-4 border-b border-zinc-800">
					<h2 class="text-lg font-semibold">Create New File</h2>
					<button
						on:click={closeNewFileDialog}
						class="p-1 hover:bg-zinc-800 rounded-lg transition-colors text-zinc-400 hover:text-zinc-200"
					>
						<X size={18} />
					</button>
				</div>
				
				<div class="p-4 space-y-4">
					<div>
						<label for="filename" class="block text-sm text-zinc-400 mb-2">File path</label>
						<input
							bind:this={newFileNameInput}
							bind:value={newFileName}
							on:keydown={handleNewFileKeydown}
							type="text"
							id="filename"
							placeholder="shell/darwin.zshrc"
							class="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 placeholder-zinc-500"
						/>
					</div>
					<p class="text-xs text-zinc-500">
						Include folder path to create nested files (e.g., <code class="text-zinc-400">shell/darwin.zshrc</code>)
					</p>
				</div>
				
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

	<div class="relative z-10 flex-1 flex flex-col overflow-hidden">
		<!-- Header -->
		<header class="flex items-center justify-between p-3 border-b border-zinc-800 bg-zinc-950/80 backdrop-blur-sm">
			<div class="flex items-center gap-3">
				<button
					on:click={goToDashboard}
					class="p-2 hover:bg-zinc-800 rounded-lg transition-colors"
					title="Back to Dashboard"
				>
					<ArrowLeft size={18} />
				</button>
				<div class="flex items-center gap-2 text-sm">
					<span class="text-zinc-500">{$auth.user?.login}/my-pact</span>
					<span class="text-zinc-600">/</span>
					<span class="text-zinc-200">{currentFile}</span>
				</div>
			</div>

			<div class="flex items-center gap-3">
				{#if saveStatus === 'saving'}
					<div class="flex items-center gap-2 text-sm text-zinc-400">
						<Loader2 size={14} class="animate-spin" />
						<span>Saving...</span>
					</div>
				{:else if saveStatus === 'saved'}
					<div class="flex items-center gap-2 text-sm text-emerald-400">
						<Check size={14} />
						<span>Saved</span>
					</div>
				{:else if saveStatus === 'error'}
					<div class="flex items-center gap-2 text-sm text-red-400">
						<span>Save failed</span>
					</div>
				{/if}

			</div>
		</header>

		{#if error}
			<div class="p-3 bg-red-500/10 border-b border-red-500/20 text-red-400 text-sm">
				{error}
			</div>
		{/if}

		{#if highlightDismissHint}
			<div class="p-2 bg-emerald-500/10 border-b border-emerald-500/20 text-emerald-400 text-xs text-center">
				Click anywhere in the editor to dismiss highlight
			</div>
		{/if}

		<div class="flex-1 flex overflow-hidden">
			<!-- Left Panel: File Tree -->
			<div class="w-64 border-r border-zinc-800 flex flex-col bg-zinc-900/30">
				<div class="flex-1 overflow-y-auto p-2">
					<!-- pact.json with sections -->
					<div class="mb-2">
						<button
							on:click={() => pactJsonExpanded = !pactJsonExpanded}
							class="w-full flex items-center gap-1 px-2 py-1 text-sm hover:bg-zinc-800/50 rounded transition-colors"
						>
							{#if pactJsonExpanded}
								<ChevronDown size={14} class="text-zinc-500" />
							{:else}
								<ChevronRight size={14} class="text-zinc-500" />
							{/if}
							<FileJson size={14} class="text-emerald-400" />
							<span class="text-zinc-200">pact.json</span>
						</button>

						{#if pactJsonExpanded}
							<div class="ml-4 border-l border-zinc-800">
								<button
									on:click={openPactJson}
									class="w-full flex items-center gap-2 px-3 py-1 text-xs hover:bg-zinc-800/50 rounded transition-colors {currentFile === 'pact.json' ? 'text-emerald-400' : 'text-zinc-400'}"
								>
									<File size={12} />
									<span>full file</span>
								</button>
								{#each pactSections as section}
									{#if section.children && section.children.length > 0}
										<!-- Section with children (e.g., ai, ricing) -->
										<div>
											<button
												on:click={() => {
													if (expandedPactSections.has(section.id)) {
														expandedPactSections.delete(section.id);
													} else {
														expandedPactSections.add(section.id);
													}
													expandedPactSections = expandedPactSections;
												}}
												class="w-full flex items-center gap-2 px-3 py-1 text-xs hover:bg-zinc-800/50 rounded transition-colors text-zinc-400 hover:text-zinc-200"
											>
												{#if expandedPactSections.has(section.id)}
													<ChevronDown size={12} />
												{:else}
													<ChevronRight size={12} />
												{/if}
												<span>{section.label}</span>
											</button>
											{#if expandedPactSections.has(section.id)}
												<div class="ml-4 border-l border-zinc-800/50">
													{#each section.children as child}
														<button
															on:click={() => highlightSection(child.id)}
															class="w-full flex items-center gap-2 px-3 py-1 text-xs hover:bg-zinc-800/50 rounded transition-colors text-zinc-500 hover:text-zinc-300"
														>
															<ChevronRight size={10} />
															<span>{child.label}</span>
														</button>
													{/each}
												</div>
											{/if}
										</div>
									{:else}
										<!-- Simple section (e.g., shell, editor) -->
										<button
											on:click={() => highlightSection(section.id)}
											class="w-full flex items-center gap-2 px-3 py-1 text-xs hover:bg-zinc-800/50 rounded transition-colors text-zinc-400 hover:text-zinc-200"
										>
											<ChevronRight size={12} />
											<span>{section.label}</span>
										</button>
									{/if}
								{/each}
							</div>
						{/if}
					</div>

					<!-- Separator -->
					<div class="h-px bg-zinc-800 my-2"></div>

					<!-- Repo files -->
					{#if repoFiles.length > 0}
						{#each repoFiles as file}
							{#if file.type === 'dir'}
								<div>
									<button
										on:click={() => toggleFolder(file.path)}
										class="w-full flex items-center gap-1 px-2 py-1 text-sm hover:bg-zinc-800/50 rounded transition-colors"
									>
										{#if expandedFolders.has(file.path)}
											<ChevronDown size={14} class="text-zinc-500" />
											<FolderOpen size={14} class="text-zinc-400" />
										{:else}
											<ChevronRight size={14} class="text-zinc-500" />
											<Folder size={14} class="text-zinc-400" />
										{/if}
										<span class="text-zinc-300">{file.name}</span>
									</button>

									{#if expandedFolders.has(file.path)}
										{#await loadFolderContents(file.path)}
											<div class="ml-6 py-1 text-xs text-zinc-500">Loading...</div>
										{:then contents}
											<div class="ml-4 border-l border-zinc-800">
												{#each contents as subFile}
													<button
														on:click={() => openFile(subFile.path)}
														class="w-full flex items-center gap-2 px-3 py-1 text-xs hover:bg-zinc-800/50 rounded transition-colors {currentFile === subFile.path ? 'text-emerald-400' : 'text-zinc-400'}"
													>
														<File size={12} />
														<span>{subFile.name}</span>
													</button>
												{/each}
											</div>
										{/await}
									{/if}
								</div>
							{:else}
								<button
									on:click={() => openFile(file.path)}
									class="w-full flex items-center gap-2 px-2 py-1 text-sm hover:bg-zinc-800/50 rounded transition-colors {currentFile === file.path ? 'text-emerald-400' : 'text-zinc-400'}"
								>
									<File size={14} />
									<span>{file.name}</span>
								</button>
							{/if}
						{/each}
					{:else}
						<div class="px-2 py-1 text-xs text-zinc-500">No files yet</div>
					{/if}
				</div>

				<!-- New file button -->
				<div class="p-2 border-t border-zinc-800">
					<button
						on:click={openNewFileDialog}
						class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 rounded-lg transition-colors"
					>
						<Plus size={14} />
						<span>New File</span>
					</button>
				</div>

				<!-- Action buttons -->
				<div class="p-2 border-t border-zinc-800 space-y-2">
					<!-- Push to GitHub -->
					<button
						on:click={saveFile}
						disabled={saving || saveStatus === 'saving'}
						class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm bg-zinc-800 border border-zinc-700 rounded-lg hover:bg-zinc-700 hover:border-zinc-600 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Github size={14} />
						<span>Push to GitHub</span>
					</button>

					<!-- Open in AI Dropdown -->
					{#if currentFile === 'pact.json'}
						<div class="relative">
							<button
								on:click|stopPropagation={toggleLlmDropdown}
								class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm bg-zinc-800 border border-zinc-700 rounded-lg hover:bg-zinc-700 hover:border-zinc-600 transition-all"
							>
								<Sparkles size={14} class="text-purple-400" />
								<span>Open in AI</span>
								<ChevronDown size={12} class="text-zinc-500" />
							</button>

							{#if showLlmDropdown}
								<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
								<div 
									class="absolute left-0 bottom-full mb-1 bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl overflow-hidden min-w-full"
									on:click|stopPropagation
								>
									<div class="p-2 border-b border-zinc-800">
										<p class="text-xs text-zinc-500">Copies config + context to clipboard</p>
									</div>
									{#each llmProviders as provider}
										<button
											on:click|stopPropagation={() => openInLlm(provider)}
											class="w-full flex items-center gap-3 px-3 py-2 hover:bg-zinc-800 transition-colors text-left"
										>
											{#if provider.id === 'chatgpt'}
												<svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
													<path d="M22.2819 9.8211a5.9847 5.9847 0 0 0-.5157-4.9108 6.0462 6.0462 0 0 0-6.5098-2.9A6.0651 6.0651 0 0 0 4.9807 4.1818a5.9847 5.9847 0 0 0-3.9977 2.9 6.0462 6.0462 0 0 0 .7427 7.0966 5.98 5.98 0 0 0 .511 4.9107 6.051 6.051 0 0 0 6.5146 2.9001A5.9847 5.9847 0 0 0 13.2599 24a6.0557 6.0557 0 0 0 5.7718-4.2058 5.9894 5.9894 0 0 0 3.9977-2.9001 6.0557 6.0557 0 0 0-.7475-7.0729zm-9.022 12.6081a4.4755 4.4755 0 0 1-2.8764-1.0408l.1419-.0804 4.7783-2.7582a.7948.7948 0 0 0 .3927-.6813v-6.7369l2.02 1.1686a.071.071 0 0 1 .038.052v5.5826a4.504 4.504 0 0 1-4.4945 4.4944zm-9.6607-4.1254a4.4708 4.4708 0 0 1-.5346-3.0137l.142.0852 4.783 2.7582a.7712.7712 0 0 0 .7806 0l5.8428-3.3685v2.3324a.0804.0804 0 0 1-.0332.0615L9.74 19.9502a4.4992 4.4992 0 0 1-6.1408-1.6464zM2.3408 7.8956a4.485 4.4850 0 0 1 2.3655-1.9728V11.6a.7664.7664 0 0 0 .3879.6765l5.8144 3.3543-2.0201 1.1685a.0757.0757 0 0 1-.071 0l-4.8303-2.7865A4.504 4.504 0 0 1 2.3408 7.8956zm16.5963 3.8558L13.1038 8.364l2.0201-1.1685a.0757.0757 0 0 1 .071 0l4.8303 2.7913a4.4944 4.4944 0 0 1-.6765 8.1042v-5.6772a.79.79 0 0 0-.407-.667zm2.0107-3.0231l-.142-.0852-4.7735-2.7818a.7759.7759 0 0 0-.7854 0L9.409 9.2297V6.8974a.0662.0662 0 0 1 .0284-.0615l4.8303-2.7866a4.4992 4.4992 0 0 1 6.6802 4.66zM8.3065 12.863l-2.02-1.1638a.0804.0804 0 0 1-.038-.0567V6.0742a4.4992 4.4992 0 0 1 7.3757-3.4537l-.142.0805L8.704 5.459a.7948.7948 0 0 0-.3927.6813zm1.0976-2.3654l2.602-1.4998 2.6069 1.4998v2.9994l-2.5974 1.4997-2.6067-1.4997Z"/>
												</svg>
											{:else}
												<img 
													src={provider.favicon} 
													alt={provider.name} 
													class="w-4 h-4 rounded"
												/>
											{/if}
											<span class="text-sm text-zinc-200">{provider.name}</span>
											<ExternalLink size={12} class="text-zinc-500 ml-auto" />
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Right Panel: Editor -->
			<div class="flex-1 flex flex-col overflow-hidden">
				{#if loading}
					<div class="flex-1 flex items-center justify-center">
						<Loader2 size={24} class="animate-spin text-zinc-500" />
					</div>
				{:else}
					<CodeEditor
						bind:this={editorComponent}
						content={currentContent}
						language={isJson ? 'json' : 'text'}
						{highlightLines}
						on:change={handleContentChange}
						on:click={handleEditorClick}
					/>
				{/if}
			</div>
		</div>
	</div>
</div>
