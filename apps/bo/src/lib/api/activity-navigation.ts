import { z } from "zod";
import type { V1ActivityRow } from "./generated/types.gen";

const id = z.string().uuid();
const searchSchema = z.object({
	from: z.string().max(10).optional(),
	to: z.string().max(10).optional(),
	actorStaffId: id.optional(),
	selected: id.optional(),
	action: z.string().max(100).optional(),
	outcome: z.enum(["success", "failure", "denied"]).optional(),
	page: z.coerce.number().int().min(1).max(1000000).default(1),
	invalid: z.boolean().optional(),
});
export function parseActivitySearch(value: unknown): z.infer<typeof searchSchema> & { invalid: boolean } {
	const parsed = searchSchema.safeParse(value);
	return parsed.success ? { ...parsed.data, invalid: parsed.data.invalid ?? false } : { page: 1, invalid: true };
}
export type ActivitySearch = ReturnType<typeof parseActivitySearch>;

export function activityReturn(search: ActivitySearch): string {
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(searchSchema.strict().parse(search))) {
		if (value !== undefined && key !== "invalid") params.set(key, String(value));
	}
	return `/activity?${params}`;
}
export function parseActivityReturn(value: string): string | undefined {
	try {
		const url = new URL(value, "https://bo.invalid");
		if (url.pathname !== "/activity" || url.hash || url.origin !== "https://bo.invalid") return;
		const raw: Record<string, string> = {};
		for (const [key, val] of url.searchParams) {
			if (key in raw || key === "invalid") return;
			raw[key] = val;
		}
		const parsed = searchSchema.strict().safeParse(raw);
		return parsed.success ? activityReturn({ ...parsed.data, invalid: false }) : undefined;
	} catch {
		return;
	}
}

export function activityFilterSearch(
	filters: Pick<ActivitySearch, "from" | "to" | "actorStaffId" | "action" | "outcome"> & Partial<Pick<ActivitySearch, "selected">>,
): ActivitySearch {
	// Controls may pass the whole URL search object despite narrower prop types.
	// Preserve the validated open detail while dropping unrelated search keys.
	return {
		from: filters.from,
		to: filters.to,
		actorStaffId: filters.actorStaffId,
		action: filters.action,
		outcome: filters.outcome,
		selected: filters.selected,
		page: 1,
		invalid: false,
	};
}

export function isActivityQuery(entry: { queryKey: readonly unknown[] }) {
	return entry.queryKey[0] === "private" && typeof entry.queryKey[3] === "string" && entry.queryKey[3].startsWith("activity");
}

// Only known UUID-owned destinations, never a server-provided URL or internal ID.
export function activityTarget(row: V1ActivityRow): string | undefined {
	if (!row.target.available || !id.safeParse(row.target.id).success) return;
	const target = row.target.id;
	if (row.target.type === "staff") return `/staff/${target}`;
	if (row.target.type === "user") return `/users/${target}`;
	if (row.target.type === "organization") return `/organizations/${target}`;
	if (row.target.type === "organization_membership")
		return id.safeParse(row.target.organizationId).success ? `/organizations/${row.target.organizationId}/members?selected=${target}` : undefined;
}
