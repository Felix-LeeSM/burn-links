import type { EncryptedTextPayload } from '$lib/crypto/text';

export const DEFAULT_API_BASE_URL =
	import.meta.env.PUBLIC_BURNLINK_API_BASE_URL || 'http://localhost:8080';

export type TTLSeconds = 600 | 3600 | 86400;

export type CreateSecretResponse = {
	id: string;
	expires_at: string;
};

export type GetSecretResponse = EncryptedTextPayload & {
	id: string;
	kind: 'text';
	expires_at: string;
};

export type ConsumeSecretResponse = {
	id: string;
	consumed: boolean;
};

export type SecretAPIClient = {
	createTextSecret(payload: EncryptedTextPayload, ttlSeconds: TTLSeconds): Promise<CreateSecretResponse>;
	getSecret(id: string): Promise<GetSecretResponse>;
	consumeSecret(id: string): Promise<ConsumeSecretResponse>;
};

type ClientOptions = {
	baseURL?: string;
	fetcher?: typeof fetch;
};

export function createSecretAPIClient(options: ClientOptions = {}): SecretAPIClient {
	const baseURL = normalizeBaseURL(options.baseURL ?? DEFAULT_API_BASE_URL);
	const fetcher = options.fetcher ?? fetch;

	return {
		async createTextSecret(payload, ttlSeconds) {
			const body = {
				kind: 'text',
				ciphertext: payload.ciphertext,
				nonce: payload.nonce,
				kdf: payload.kdf,
				size_bytes: payload.size_bytes,
				ttl_seconds: ttlSeconds,
				max_views: 1
			};

			return requestJSON<CreateSecretResponse>(fetcher, `${baseURL}/api/secrets`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});
		},

		getSecret(id) {
			return requestJSON<GetSecretResponse>(fetcher, `${baseURL}/api/secrets/${encodeURIComponent(id)}`);
		},

		consumeSecret(id) {
			return requestJSON<ConsumeSecretResponse>(fetcher, `${baseURL}/api/secrets/${encodeURIComponent(id)}/consume`, {
				method: 'POST'
			});
		}
	};
}

export function createShareURL(origin: string, id: string): string {
	const url = new URL(origin);
	url.pathname = `/s/${encodeURIComponent(id)}`;
	url.search = '';
	url.hash = '';
	return url.toString();
}

async function requestJSON<T>(fetcher: typeof fetch, input: RequestInfo | URL, init?: RequestInit): Promise<T> {
	const response = await fetcher(input, init);
	if (!response.ok) {
		throw new Error(await errorMessage(response));
	}
	return (await response.json()) as T;
}

async function errorMessage(response: Response): Promise<string> {
	try {
		const body = (await response.json()) as { error?: { message?: string } };
		return body.error?.message ?? `Request failed with status ${response.status}`;
	} catch {
		return `Request failed with status ${response.status}`;
	}
}

function normalizeBaseURL(value: string): string {
	return value.replace(/\/+$/, '');
}
