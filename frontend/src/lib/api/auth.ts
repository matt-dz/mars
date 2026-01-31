import { type FetchFn } from '@/http';
import { isHTTPError } from 'ky';
import { ApiErrorSchema, HTTPError } from './errors';

export async function verifySession(fetch: FetchFn) {
	try {
		return await fetch.get('api/auth/verify');
	} catch (e) {
		if (isHTTPError(e)) {
			const err = ApiErrorSchema.safeParse(await e.response.clone().json());
			if (err.success) {
				throw new HTTPError(err.data.status, err.data.message, err.data.code, err.data.error_id);
			}
			throw new HTTPError(e.response.status, await e.response.text());
		}
		throw e;
	}
}

export async function refreshSession(fetch: FetchFn) {
	try {
		return await fetch.post('api/auth/refresh', { json: {} });
	} catch (e) {
		if (isHTTPError(e)) {
			const err = ApiErrorSchema.safeParse(await e.response.clone().json());
			if (err.success) {
				throw new HTTPError(err.data.status, err.data.message, err.data.code, err.data.error_id);
			}
			throw new HTTPError(e.response.status, await e.response.text());
		}
		throw e;
	}
}
