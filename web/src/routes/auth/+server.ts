import { redirect } from '@sveltejs/kit';
import { GITHUB_CLIENT_ID } from '$lib/github';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	const params = new URLSearchParams({
		client_id: GITHUB_CLIENT_ID,
		scope: 'repo'
	});

	throw redirect(302, `https://github.com/login/oauth/authorize?${params}`);
};
