import { json, error } from '@sveltejs/kit';
import { GITHUB_CLIENT_ID } from '$lib/github';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, platform }) => {
	const { code } = await request.json();

	if (!code) {
		throw error(400, 'Missing authorization code');
	}

	// Get client secret from Cloudflare environment
	// Try multiple ways to access it
	const GITHUB_CLIENT_SECRET = 
		platform?.env?.GITHUB_CLIENT_SECRET || 
		(globalThis as any).GITHUB_CLIENT_SECRET ||
		'';

	if (!GITHUB_CLIENT_SECRET) {
		console.error('GITHUB_CLIENT_SECRET not found. Platform:', JSON.stringify(platform));
		throw error(500, 'Server configuration error: missing client secret');
	}

	try {
		// Exchange code for access token
		const tokenResponse = await fetch('https://github.com/login/oauth/access_token', {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				client_id: GITHUB_CLIENT_ID,
				client_secret: GITHUB_CLIENT_SECRET,
				code
			})
		});

		const tokenData = await tokenResponse.json();

		if (tokenData.error) {
			console.error('GitHub OAuth error:', tokenData.error, tokenData.error_description);
			throw error(400, tokenData.error_description || tokenData.error);
		}

		if (!tokenData.access_token) {
			console.error('No access token in response:', JSON.stringify(tokenData));
			throw error(500, 'No access token received from GitHub');
		}

		return json({ access_token: tokenData.access_token });
	} catch (e) {
		console.error('Token exchange error:', e);
		if (e && typeof e === 'object' && 'status' in e) {
			throw e; // Re-throw SvelteKit errors
		}
		throw error(500, 'Failed to exchange code for token');
	}
};
