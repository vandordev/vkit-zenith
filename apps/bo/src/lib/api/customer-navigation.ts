import { z } from "zod";
import * as schema from "./generated/zod.gen";
const page = z.coerce.number().int().min(1).max(1000000).catch(1);
const q = z.string().max(100).catch("");
const selected = z.string().uuid().optional().catch(undefined);
export const organizationSections = ["members", "api-keys", "activity"] as const;
export const organizationSearch = z.object({
	q,
	status: schema.zBoV1ListOrganizationsQuery.shape.status.catch("").transform((v) => v ?? ""),
	page,
	create: z.boolean().optional().catch(undefined),
});
export const userSearch = z.object({
	q,
	status: schema.zBoV1ListUsersQuery.shape.status.catch("").transform((v) => v ?? ""),
	verification: schema.zBoV1ListUsersQuery.shape.verification.catch("").transform((v) => v ?? ""),
	page,
	create: z.boolean().optional().catch(undefined),
});
const returnTo = z
	.unknown()
	.optional()
	.transform((value) => parseCustomerReturn(value));
export const organizationOverviewSearch = z.object({
	action: z.enum(["profile", "status"]).optional().catch(undefined),
	returnTo,
});
export const organizationMembersSearch = z
	.object({
		q,
		status: schema.zBoV1ListOrganizationMembershipsQuery.shape.status.catch("").transform((v) => v ?? ""),
		ownership: schema.zBoV1ListOrganizationMembershipsQuery.shape.ownership.catch("").transform((v) => v ?? ""),
		page,
		selected,
		action: z.enum(["add", "owner"]).optional().catch(undefined),
		returnTo,
	})
	.transform((value) => (value.action === "owner" && !value.selected ? { ...value, action: undefined } : value));
// Organization sections page only.
export const organizationSectionSearch = z.object({ page });
// User context sub-routes (T10): `organizations` pages memberships and keeps the validated origin;
// sessions and activity page only.
export const userSections = ["organizations", "sessions", "activity"] as const;
export const userOrganizationsSearch = z.object({ page, returnTo });
export const userSessionsSearch = z.object({ page, returnTo });
export const userActivitySearch = z.object({ q, page, returnTo });
const staffStatus = z.enum(["", "active", "disabled"]);
const staffReturnState = z.object({ q: q.default(""), status: staffStatus.catch("").default(""), page });
export const staffDetailSearch = z.object({ returnTo: z.unknown().optional().transform(parseStaffReturn) });
export const staffSectionSearch = z.object({ page, returnTo: z.unknown().optional().transform(parseStaffReturn) });
export const staffActivitySearch = z.object({ q, page, returnTo: z.unknown().optional().transform(parseStaffReturn) });

export function staffListReturn(search: { q?: string; status?: string; page?: number }): string {
	const parsed = staffReturnState.safeParse(search);
	const state = parsed.success ? parsed.data : { q: "", status: "" as const, page: 1 };
	const params = new URLSearchParams();
	if (state.q) params.set("q", state.q);
	if (state.status) params.set("status", state.status);
	params.set("page", String(state.page));
	return `/staff?${params}`;
}

export function parseStaffReturn(value: unknown): string | undefined {
	if (typeof value !== "string" || value.length > 2000 || !value.startsWith("/staff?") || value.startsWith("//")) return;
	try {
		const url = new URL(value, "https://bo.invalid");
		if (url.origin !== "https://bo.invalid" || url.pathname !== "/staff" || url.hash) return;
		const raw: Record<string, string> = {};
		for (const [key, entry] of url.searchParams) {
			if (key in raw) return;
			raw[key] = entry;
		}
		const parsed = staffReturnState.strict().safeParse(raw);
		return parsed.success ? staffListReturn(parsed.data) : undefined;
	} catch {
		return;
	}
}

export function staffReturnSearch(value: unknown) {
	const parsed = parseStaffReturn(value);
	const params = parsed ? new URL(parsed, "https://bo.invalid").searchParams : undefined;
	return staffReturnState.parse({ q: params?.get("q") ?? undefined, status: params?.get("status") ?? undefined, page: params?.get("page") ?? undefined });
}
export const userDetailSearch = z.object({
	page,
	action: z.enum(["profile", "email", "verify", "status"]).optional().catch(undefined),
	returnTo,
});

// At most one return to a known customer page. No arbitrary URLs, hash, nested
// return chains, secret payloads or silent acceptance of malformed filter values.
export function parseCustomerReturn(value: unknown): string | undefined {
	return parseCustomerOnlyReturn(value);
}
export function parseCustomerOnlyReturn(value: unknown): string | undefined {
	return parseReturn(value, true);
}
function parseReturn(value: unknown, allowOrigin: boolean): string | undefined {
	if (
		typeof value !== "string" ||
		value.length > 2000 ||
		!value.startsWith("/") ||
		value.startsWith("//") ||
		/[\\\u0000-\u0020\u007f]/u.test(value)
	)
		return;
	try {
		const url = new URL(value, "https://bo.invalid");
		if (
			url.origin !== "https://bo.invalid" ||
			url.hash ||
			url.pathname !== value.split(/[?#]/)[0] ||
			/[\\\u0000-\u001f\u007f]/u.test(decodeURIComponent(url.search))
		)
			return;
		const segments = url.pathname.split("/").filter(Boolean);
		if (segments.length < 1 || segments.length > 3 || !["organizations", "users"].includes(segments[0] ?? "")) return;
		const detail = segments.length >= 2;
		if (detail && !z.string().uuid().safeParse(segments[1]).success) return;
		// Organization context sub-routes: members carries list state, the rest only `page`.
		const section = detail ? segments[2] : undefined;
		if (segments[0] === "users" && segments.length > 2 && (segments.length > 3 || !(userSections as readonly string[]).includes(segments[2] ?? ""))) return;
		if (segments[0] === "organizations" && segments.length >= 3 && !(organizationSections as readonly string[]).includes(section ?? "")) return;
		const memberList = segments[0] === "organizations" && section === "members";
		const output = new URLSearchParams();
		const seen = new Set<string>();
		for (const [key, val] of url.searchParams) {
			if (seen.has(key)) return;
			seen.add(key);
			if (key === "returnTo") {
				if (!detail || !allowOrigin) return;
				const origin = parseReturn(val, false);
				if (!origin) return;
				output.set(key, origin);
				continue;
			}
			if (key === "page") {
				if (!/^[1-9][0-9]{0,6}$/.test(val) || Number(val) > 1000000) return;
				// Pages exist on members and on every list/detail section, not on the overview.
				if (segments[0] === "organizations" && detail && segments.length < 3) return;
			} else if (key === "q") {
				const userActivity = segments[0] === "users" && section === "activity";
				if ((detail && !memberList && !userActivity) || val.length > 100) return;
			} else if (key === "status") {
				const values =
					segments[0] === "users" ? ["", "active", "disabled"] : detail ? ["", "active", "removed"] : ["", "active", "suspended"];
				if (!values.includes(val) || (detail && !memberList)) return;
			} else if (key === "verification") {
				if (detail || segments[0] !== "users" || !["", "verified", "unverified"].includes(val)) return;
			} else if (key === "ownership") {
				if (!memberList || !["", "owner", "member"].includes(val)) return;
			} else if (key === "selected") {
				if (!memberList || !z.string().uuid().safeParse(val).success) return;
			} else return;
			output.set(key, val);
		}
		const suffix = output.toString();
		return `/${segments.join("/")}${suffix ? `?${suffix}` : ""}`;
	} catch {
		return;
	}
}
export function customerReturn(path: string, search: Record<string, unknown>): string {
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(search)) {
		if (value !== undefined && key !== "create" && key !== "action") params.set(key, String(value));
	}
	return parseCustomerReturn(`${path}?${params}`) ?? (path.startsWith("/organizations") ? "/organizations" : "/users");
}
export function customerOrigin(path: string, search: Record<string, unknown>): string {
	// Keep the original validated origin through cross-context jumps, never append
	// another detail to an existing nested return chain.
	return parseCustomerReturn(search.returnTo) ?? customerReturn(path, { ...search, returnTo: undefined });
}

