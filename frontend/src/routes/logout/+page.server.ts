import {
	ACCESS_TOKEN_COOKIE_NAME,
	CSRF_TOKEN_COOKIE_NAME,
	REFRESH_TOKEN_COOKIE_NAME
} from '@/auth';
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies }) => {
	cookies.delete(ACCESS_TOKEN_COOKIE_NAME, { path: '/' });
	cookies.delete(REFRESH_TOKEN_COOKIE_NAME, { path: '/' });
	cookies.delete(CSRF_TOKEN_COOKIE_NAME, { path: '/' });
	redirect(302, '/login');
};
