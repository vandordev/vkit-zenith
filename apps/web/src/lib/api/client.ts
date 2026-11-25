import type { PublicWebConfig } from "@repo/config/public";
import { z } from "zod";
import { createClient, createConfig } from "./generated/client";
import { authV1GetCsrf } from "./generated/sdk.gen";
import { AuthError } from "./errors";

// Core signs CSRF for 10 minutes (identity/service.go newCSRF); refresh one minute early.
const csrfLifetime = 9 * 60 * 1000;
const problemSchema = z.object({
	type: z.string().optional(),
	detail: z.string().optional(),
	errors: z.array(z.object({ message: z.string().optional(), location: z.string().optional() })).optional(),
});

export function createAuthTransport(config: PublicWebConfig, onIdentityChange: () => void, fetcher: typeof fetch = fetch) {
	let epoch = 0;
	let cached: { token: string; until: number } | undefined;
	let acquiring: Promise<string> | undefined;
	const client = createClient(
		createConfig({
			baseUrl: config.apiBaseUrl,
			credentials: "include",
			cache: "no-store",
			fetch: Object.assign(
				async (input: Parameters<typeof fetch>[0], init?: Parameters<typeof fetch>[1]) => {
					const accepted = epoch;
					const received = await fetcher(input, init);
					// Read the whole body here so the epoch fence covers it; the SDK parses the buffered copy.
					const bytes = await received.arrayBuffer();
					if (accepted !== epoch) throw new AuthError(401, "session_changed", "Sesi berubah.");
					const response = new Response(bytes.byteLength ? bytes : null, received);
					if (response.ok) return response;
					if (response.status === 403) cached = undefined;
					let parsed: ReturnType<typeof problemSchema.safeParse>;
					try {
						parsed = problemSchema.safeParse(JSON.parse(new TextDecoder().decode(bytes)));
					} catch {
						parsed = problemSchema.safeParse(null);
					}
					const problem = parsed.success ? parsed.data : {};
					const fields: Partial<Record<string, string>> = {};
					for (const item of problem.errors ?? []) {
						const name = item.location?.replace(/^body\./, "");
						if (name && item.message) fields[name] = item.message;
					}
					const retry = Number(response.headers.get("Retry-After"));
					throw new AuthError(
						response.status,
						problem.type ?? "unknown",
						problem.detail ?? "Permintaan ditolak.",
						Number.isFinite(retry) && retry > 0 ? retry : undefined,
						Object.keys(fields).length ? fields : undefined,
					);
				},
				{ preconnect: fetcher.preconnect },
			),
		}),
	);
	function setCsrf(token: string) {
		cached = { token, until: Date.now() + csrfLifetime };
	}
	async function csrf(): Promise<string> {
		if (cached && Date.now() < cached.until) return cached.token;
		if (acquiring) return acquiring;
		const accepted = epoch;
		const pending = authV1GetCsrf({ client, throwOnError: true }).then(({ data }) => {
			if (accepted !== epoch) throw new AuthError(401, "session_changed", "Sesi berubah.");
			setCsrf(data.token);
			return data.token;
		});
		acquiring = pending;
		try {
			return await pending;
		} finally {
			if (acquiring === pending) acquiring = undefined;
		}
	}
	return {
		client,
		csrf,
		setCsrf,
		epoch: () => epoch,
		identityChanged() {
			epoch++;
			cached = undefined;
			acquiring = undefined;
			onIdentityChange();
		},
	};
}
export type AuthTransport = ReturnType<typeof createAuthTransport>;
