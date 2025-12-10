<script lang="ts">
  import { Settings, Terminal, Key, Check, AlertCircle, ChevronRight, Cpu, Palette, MessageSquare, FolderGit2, RefreshCw } from 'lucide-svelte';

  const kitModules = [
    { id: 'shell', name: 'Shell', icon: Terminal, useCustomIcon: false, status: 'synced', files: 3, description: 'zsh + oh-my-posh' },
    { id: 'editor', name: 'Editor', icon: Cpu, useCustomIcon: false, status: 'synced', files: 12, description: 'neovim + lazy.nvim' },
    { id: 'terminal', name: 'Terminal', icon: Terminal, useCustomIcon: false, status: 'synced', files: 1, description: 'ghostty + catppuccin' },
    { id: 'git', name: 'Git', icon: null, useCustomIcon: true, status: 'synced', files: 2, description: 'aliases + config' },
    { id: 'ai', name: 'AI', icon: MessageSquare, useCustomIcon: false, status: 'pending', files: 8, description: '3 providers, 5 prompts' },
    { id: 'tools', name: 'Tools', icon: Cpu, useCustomIcon: false, status: 'synced', files: 4, description: 'lazygit, ripgrep, fzf' },
    { id: 'keybindings', name: 'Keybindings', icon: Cpu, useCustomIcon: false, status: 'synced', files: 2, description: 'vscode + nvim' },
    { id: 'snippets', name: 'Snippets', icon: Cpu, useCustomIcon: false, status: 'not_configured', files: 0, description: 'not configured' },
    { id: 'fonts', name: 'Fonts', icon: Palette, useCustomIcon: false, status: 'missing', files: 1, description: '1 missing' },
  ];

  const secrets = [
    { name: 'ANTHROPIC_API_KEY', set: true, lastUsed: '2 hours ago' },
    { name: 'OPENAI_API_KEY', set: true, lastUsed: '1 day ago' },
    { name: 'GITHUB_TOKEN', set: true, lastUsed: '5 mins ago' },
    { name: 'GROQ_API_KEY', set: false, lastUsed: null },
  ];

  const recentActivity = [
    { action: 'Synced shell config', time: '5 mins ago' },
    { action: 'Updated nvim plugins', time: '2 hours ago' },
    { action: 'Added ANTHROPIC_API_KEY', time: '1 day ago' },
    { action: 'Pushed to remote', time: '2 days ago' },
  ];

  function getStatusColor(status: string) {
    switch (status) {
      case 'synced': return 'text-emerald-400';
      case 'pending': return 'text-amber-400';
      case 'missing': return 'text-amber-400';
      case 'not_configured': return 'text-zinc-500';
      default: return 'text-zinc-500';
    }
  }

  function getStatusIcon(status: string) {
    switch (status) {
      case 'synced': return Check;
      case 'pending': return AlertCircle;
      case 'missing': return AlertCircle;
      default: return null;
    }
  }

  function getStatusText(status: string) {
    switch (status) {
      case 'synced': return 'synced';
      case 'pending': return 'pending';
      case 'missing': return 'missing';
      case 'not_configured': return 'not configured';
      default: return status;
    }
  }
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100 font-mono">
  <!-- Subtle grid background -->
  <div class="fixed inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none"></div>
  
  <div class="relative z-10 max-w-6xl mx-auto p-8">
    <!-- Header -->
    <header class="flex items-center justify-between mb-12">
      <div class="flex items-center gap-4">
        <div class="w-10 h-10 bg-gradient-to-br from-emerald-400 to-emerald-600 rounded-lg flex items-center justify-center">
          <FolderGit2 size={20} class="text-zinc-950" />
        </div>
        <div>
          <h1 class="text-xl font-bold tracking-tight">pact</h1>
          <p class="text-xs text-zinc-500">cloudboy-jh/pact</p>
        </div>
      </div>
      
      <div class="flex items-center gap-3">
        <button class="flex items-center gap-2 px-4 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-sm hover:bg-zinc-800 hover:border-zinc-700 transition-all">
          <RefreshCw size={14} />
          <span>Sync</span>
        </button>
        <button class="p-2 bg-zinc-900 border border-zinc-800 rounded-lg hover:bg-zinc-800 transition-all">
          <Settings size={16} />
        </button>
      </div>
    </header>

    <!-- Status bar -->
    <div class="flex items-center gap-6 mb-8 p-4 bg-zinc-900/50 border border-zinc-800/50 rounded-xl">
      <div class="flex items-center gap-2">
        <div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
        <span class="text-sm text-zinc-400">All synced</span>
      </div>
      <div class="h-4 w-px bg-zinc-800"></div>
      <span class="text-sm text-zinc-500">Last sync: 5 mins ago</span>
      <div class="h-4 w-px bg-zinc-800"></div>
      <span class="text-sm text-zinc-500">Machine: <span class="text-zinc-300">macbook-pro</span></span>
    </div>

    <div class="grid grid-cols-3 gap-6">
      <!-- Kit Modules -->
      <div class="col-span-2 space-y-4">
        <h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider mb-4">Your Kit</h2>
        
        <div class="space-y-2">
          {#each kitModules as module}
            <div
              class="group flex items-center justify-between p-4 bg-zinc-900/30 border border-zinc-800/50 rounded-xl hover:bg-zinc-900/60 hover:border-zinc-700/50 transition-all cursor-pointer"
            >
              <div class="flex items-center gap-4">
                <div class="w-10 h-10 bg-zinc-800 rounded-lg flex items-center justify-center group-hover:bg-zinc-700 transition-colors">
                  {#if module.useCustomIcon}
                    <img src="/pixel-head-white.png" alt={module.name} class="w-5 h-5" />
                  {:else}
                    <svelte:component this={module.icon} size={18} class="text-zinc-400" />
                  {/if}
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <span class="font-medium">{module.name}</span>
                    {#if module.files > 0}
                      <span class="text-xs text-zinc-600">{module.files} files</span>
                    {/if}
                  </div>
                  <span class="text-sm text-zinc-500">{module.description}</span>
                </div>
              </div>
              
              <div class="flex items-center gap-3">
                <span class="flex items-center gap-1.5 text-xs {getStatusColor(module.status)}">
                  {#if getStatusIcon(module.status)}
                    <svelte:component this={getStatusIcon(module.status)} size={12} />
                  {/if}
                  {getStatusText(module.status)}
                </span>
                <ChevronRight size={16} class="text-zinc-600 group-hover:text-zinc-400 transition-colors" />
              </div>
            </div>
          {/each}
        </div>

        <!-- Add module button -->
        <button class="w-full p-4 border border-dashed border-zinc-800 rounded-xl text-zinc-500 hover:border-zinc-600 hover:text-zinc-400 transition-all">
          + Add module
        </button>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Secrets -->
        <div>
          <h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider mb-4 flex items-center gap-2">
            <Key size={14} />
            Secrets
          </h2>
          
          <div class="bg-zinc-900/30 border border-zinc-800/50 rounded-xl overflow-hidden">
            {#each secrets as secret, i}
              <div
                class="flex items-center justify-between p-3 {i !== secrets.length - 1 ? 'border-b border-zinc-800/50' : ''}"
              >
                <div class="flex items-center gap-2">
                  {#if secret.set}
                    <div class="w-1.5 h-1.5 bg-emerald-400 rounded-full"></div>
                  {:else}
                    <div class="w-1.5 h-1.5 bg-zinc-600 rounded-full"></div>
                  {/if}
                  <span class="text-xs font-mono text-zinc-300">{secret.name}</span>
                </div>
                {#if secret.set}
                  <span class="text-xs text-zinc-600">{secret.lastUsed}</span>
                {:else}
                  <button class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors">set</button>
                {/if}
              </div>
            {/each}
          </div>
        </div>

        <!-- Recent Activity -->
        <div>
          <h2 class="text-sm font-medium text-zinc-400 uppercase tracking-wider mb-4">Activity</h2>
          
          <div class="space-y-3">
            {#each recentActivity as activity}
              <div class="flex items-start gap-3">
                <div class="w-1 h-1 bg-zinc-600 rounded-full mt-2"></div>
                <div>
                  <p class="text-sm text-zinc-300">{activity.action}</p>
                  <p class="text-xs text-zinc-600">{activity.time}</p>
                </div>
              </div>
            {/each}
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
      </div>
    </div>
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>
