import { json, error } from '@sveltejs/kit';
import { GITHUB_CLIENT_ID } from '$lib/github';
import type { RequestHandler } from './$types';

// In production, this should be an environment variable
const GITHUB_CLIENT_SECRET = process.env.GITHUB_CLIENT_SECRET || 'YOUR_GITHUB_CLIENT_SECRET';

export const POST: RequestHandler = async ({ request }) => {
	const { code } = await request.json();

	if (!code) {
		throw error(400, 'Missing authorization code');
	}

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

	if (!tokenResponse.ok) {
		throw error(500, 'Failed to exchange code for token');
	}

	const tokenData = await tokenResponse.json();

	if (tokenData.error) {
		throw error(400, tokenData.error_description || tokenData.error);
	}

	return json({ access_token: tokenData.access_token });
};
