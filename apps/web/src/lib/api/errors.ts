import type { FormStatus } from "../../app/-components/form-status";

export class AuthError extends Error {
	constructor(
		readonly status: number,
		readonly code: string,
		message: string,
		readonly retryAfter?: number,
		readonly fields?: Partial<Record<string, string>>,
	) {
		super(message);
	}
}

const messages: Partial<Record<number, string>> = {
	401: "Email atau kata sandi salah, atau sesi Anda berakhir.",
	403: "Konfirmasi keamanan kedaluwarsa. Periksa isian lalu kirim kembali.",
	404: "Permintaan tidak ditemukan atau sudah kedaluwarsa.",
};

export function toFormStatus(error: unknown): FormStatus {
	if (!(error instanceof AuthError)) return { kind: "unavailable" };
	if (error.status === 422) return { kind: "error", message: "Periksa kembali isian Anda.", ...(error.fields && { fields: error.fields }) };
	if (error.status === 429) return { kind: "rate-limited", retryAfterSeconds: error.retryAfter };
	const message = messages[error.status];
	return message ? { kind: "error", message } : { kind: "unavailable" };
}
