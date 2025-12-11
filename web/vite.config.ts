import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	// Load env from parent directory (.env.local at project root)
	const env = loadEnv(mode, process.cwd() + '/..', '');
	
	return {
		plugins: [sveltekit()],
		define: {
			// Make GITHUB_CLIENT_SECRET available server-side
			'process.env.GITHUB_CLIENT_SECRET': JSON.stringify(env.GITHUB_CLIENT_SECRET),
			// Make GITHUB_CLIENT_ID available client-side
			'process.env.GITHUB_CLIENT_ID': JSON.stringify(env.GITHUB_CLIENT_ID)
		}
	};
});
