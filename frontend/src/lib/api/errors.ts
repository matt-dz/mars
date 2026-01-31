import * as z from 'zod';

/**
 * Error codes matching the backend api/internal/api/error/codes.go
 */
export const ErrorCode = {
	UnknownError: 'unknown_error',
	InternalServerError: 'internal_server_error',
	BadRequest: 'bad_request',
	UnprocessableEntity: 'unprocessible_entity',
	InvalidCredentials: 'invalid_credentials',
	InvalidAccessToken: 'invalid_access_token',
	ExpiredAccessToken: 'expired_access_token',
	InvalidRefreshToken: 'invalid_refresh_token',
	ExpiredRefreshToken: 'expired_refresh_token',
	InsufficientPermissions: 'insufficient_permissions'
} as const;

export type ErrorCode = (typeof ErrorCode)[keyof typeof ErrorCode];

/**
 * Error codes that indicate the refresh token is invalid/expired.
 * When these are returned, we cannot recover and must redirect to login.
 */
export const UNRECOVERABLE_AUTH_CODES: ErrorCode[] = [
	ErrorCode.InvalidRefreshToken,
	ErrorCode.ExpiredRefreshToken
];

/**
 * Error codes that indicate the access token is invalid/expired.
 * These can potentially be recovered by refreshing the token.
 */
export const RECOVERABLE_AUTH_CODES: ErrorCode[] = [
	ErrorCode.InvalidAccessToken,
	ErrorCode.ExpiredAccessToken
];

export const ApiErrorSchema = z.object({
	code: z.string(),
	error_id: z.number(),
	message: z.string(),
	status: z.number()
});

export type ApiError = z.infer<typeof ApiErrorSchema>;

/**
 * Schema for login/refresh response matching backend LoginResponse
 */
export const LoginResponseSchema = z.object({
	access_token: z.string(),
	token_type: z.string(),
	expires_in: z.number()
});

export type LoginResponse = z.infer<typeof LoginResponseSchema>;

/**
 * Checks if an error code indicates an unrecoverable auth failure
 */
export function isUnrecoverableAuthError(code: string): boolean {
	return UNRECOVERABLE_AUTH_CODES.includes(code as ErrorCode);
}

/**
 * Checks if an error code indicates a recoverable auth failure (can refresh)
 */
export function isRecoverableAuthError(code: string): boolean {
	return RECOVERABLE_AUTH_CODES.includes(code as ErrorCode);
}

/**
 * Custom error class for authentication failures that require redirect
 */
export class AuthenticationError extends Error {
	public readonly code: ErrorCode | string;
	public readonly shouldRedirect: boolean;

	constructor(message: string, code: ErrorCode | string, shouldRedirect: boolean = true) {
		super(message);
		this.name = 'AuthenticationError';
		this.code = code;
		this.shouldRedirect = shouldRedirect;
	}
}

export class HTTPError extends Error {
	public readonly statusCode: number;
	public readonly message: string;
	public readonly errorCode: ErrorCode | string;
	public readonly requestId: number;

	constructor(statusCode: number, message: string, errorCode: string = '', requestId: number = 0) {
		super(
			`HTTP Status Error: statusCode=${statusCode} errorCode=${errorCode} requestId=${requestId} body="${message}"`
		);
		this.name = 'HTTPError';
		this.statusCode = statusCode;
		this.message = message;
		this.errorCode = errorCode;
		this.requestId = requestId;
	}
}
