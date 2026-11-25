import type { PublicBOConfig } from "@repo/config/public";
import { z } from "zod";
import { createClient, createConfig } from "./generated/client";
import { boV1GetCsrf } from "./generated/sdk.gen";

const problemSchema = z.object({ type: z.string().optional(), detail: z.string().optional() });
export class BOError extends Error {
	constructor(
		readonly status: number,
		readonly code: string,
		detail: string,
		readonly retryAfter?: string,
	) {
		super(detail);
	}
}
export function errorMessage(error: unknown): string {
	if (error instanceof BOError) {
		if (error.status === 409) return `${error.message} Muat ulang data terbaru sebelum menyimpan. Isian Anda tetap tersedia.`;
		if (error.status === 429) return `${error.message} Tunggu ${error.retryAfter ?? "beberapa"} detik sebelum mencoba lagi.`;
		if (error.status === 403)
			return `${error.message} Periksa hak akses Anda. Jika konfirmasi keamanan berubah atau kedaluwarsa, periksa isian lalu kirim kembali. Tidak ada pengiriman ulang otomatis.`;
		if (error.code === "csrf")
			return "Konfirmasi keamanan kedaluwarsa atau berubah di tab lain. Coba kirim kembali setelah memeriksa isian. Tidak ada pengiriman ulang otomatis.";
		return error.message;
	}
	return "Permintaan tidak dapat diselesaikan. Periksa koneksi dan coba lagi.";
}

export function createBOTransport(config: PublicBOConfig, lost: () => void, fetcher: typeof fetch = fetch) {
	let generation = 0;
	let csrf: { token: string; until: number } | undefined;
	let acquiring: Promise<string> | undefined;
	function reset() {
		generation++;
		csrf = undefined;
		acquiring = undefined;
		lost();
	}
	const client = createClient(
		createConfig({
			baseUrl: config.boApiBaseUrl,
			credentials: "include",
			cache: "no-store",
			fetch: Object.assign(
				async (input: Parameters<typeof fetch>[0], init?: Parameters<typeof fetch>[1]) => {
					const accepted = generation;
					const response = await fetcher(input, init);
					if (accepted !== generation) throw new BOError(401, "session_changed", "Sesi berubah. Silakan masuk kembali.");
					if (response.status === 401) reset();
					if (!response.ok) {
						const parsed = problemSchema.safeParse(
							await response
								.clone()
								.json()
								.catch(() => null),
						);
						const code = parsed.success ? (parsed.data.type?.replace("urn:zenith:bo:", "") ?? "unknown") : "unknown";
						if (response.status === 403) csrf = undefined;
						throw new BOError(
							response.status,
							code,
							parsed.success ? (parsed.data.detail ?? "Permintaan ditolak.") : "Permintaan ditolak.",
							response.headers.get("Retry-After") ?? undefined,
						);
					}
					return response;
				},
				{ preconnect: fetcher.preconnect },
			),
		}),
	);
	async function token(): Promise<string> {
		if (csrf && Date.now() < csrf.until) return csrf.token;
		if (acquiring) return acquiring;
		const accepted = generation;
		const pending = boV1GetCsrf({ client, throwOnError: true }).then(({ data }) => {
			if (accepted !== generation) throw new BOError(401, "session_changed", "Sesi berubah.");
			csrf = { token: data.token, until: Date.now() + 29 * 60 * 1000 };
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
		reset,
		// Always explicit. An unsafe write is never replayed after CSRF/session failure.
		async writeOptions() {
			return { client, throwOnError: true as const, headers: { "Content-Type": "application/json", "X-CSRF-Token": await token() } };
		},
		rotateCSRF() {
			csrf = undefined;
			acquiring = undefined;
		},
	};
}
export type BOTransport = ReturnType<typeof createBOTransport>;
