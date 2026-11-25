import { Badge } from "@repo/ui/badge";

type Tone = "success" | "warning" | "info" | "destructive" | "neutral";
const variants = {
	success: "success-light",
	warning: "warning-light",
	info: "info-light",
	destructive: "destructive-light",
	neutral: "outline",
} as const satisfies Record<Tone, string>;

// One map for every schema status the Backoffice renders. Unknown values from a newer
// API fall back to a neutral badge with the raw value, never an exception.
const statuses = {
	apiKey: { active: ["Aktif", "success"], expired: ["Kedaluwarsa", "warning"], revoked: ["Dicabut", "destructive"] },
	job: {
		available: ["Tersedia", "info"],
		scheduled: ["Terjadwal", "info"],
		running: ["Berjalan", "info"],
		retryable: ["Akan diulang", "warning"],
		completed: ["Selesai", "success"],
		cancelled: ["Dibatalkan", "neutral"],
		discarded: ["Dibuang", "destructive"],
	},
	session: { active: ["Aktif", "success"], expired: ["Kedaluwarsa", "neutral"], revoked: ["Dicabut", "destructive"] },
	emailAction: {
		pending: ["Menunggu", "warning"],
		consumed: ["Digunakan", "success"],
		revoked: ["Dicabut", "destructive"],
		expired: ["Kedaluwarsa", "neutral"],
	},
	audit: { success: ["Berhasil", "success"], failure: ["Gagal", "destructive"], denied: ["Ditolak", "warning"] },
} as const satisfies Record<string, Record<string, readonly [string, Tone]>>;

export type StatusKind = keyof typeof statuses;

export function statusPresentation(kind: StatusKind, status: string): { label: string; tone: Tone } {
	const known = (statuses[kind] as Record<string, readonly [string, Tone]>)[status];
	return known ? { label: known[0], tone: known[1] } : { label: status, tone: "neutral" };
}

export function StatusBadge({ kind, status }: { kind: StatusKind; status: string }) {
	const { label, tone } = statusPresentation(kind, status);
	return (
		<Badge variant={variants[tone]} data-status={status} className="max-w-full">
			<span className="truncate">{label}</span>
		</Badge>
	);
}
