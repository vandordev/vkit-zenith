import { Badge } from "@repo/ui/badge";

type Tone = "success" | "warning" | "info" | "destructive" | "neutral";
const variants = {
	success: "success-light",
	warning: "warning-light",
	info: "info-light",
	destructive: "destructive-light",
	neutral: "outline",
} as const satisfies Record<Tone, string>;

// resource → status → {label, tone}. Later menus extend this map. An unknown value from a newer API
// renders neutral with the raw value, never an exception.
const statuses = {
	organization: { active: ["Aktif", "success"], suspended: ["Ditangguhkan", "destructive"] },
	apiKey: { active: ["Aktif", "success"], expired: ["Kedaluwarsa", "neutral"], revoked: ["Dicabut", "destructive"] },
} as const satisfies Record<string, Record<string, readonly [string, Tone]>>;

export type StatusResource = keyof typeof statuses;

export function statusPresentation(resource: StatusResource, status: string): { label: string; tone: Tone } {
	const known = (statuses[resource] as Record<string, readonly [string, Tone]>)[status];
	return known ? { label: known[0], tone: known[1] } : { label: status, tone: "neutral" };
}

export function StatusBadge({ resource, status }: { resource: StatusResource; status: string }) {
	const { label, tone } = statusPresentation(resource, status);
	return (
		<Badge variant={variants[tone]} data-status={status} className="max-w-full">
			<span className="truncate">{label}</span>
		</Badge>
	);
}
