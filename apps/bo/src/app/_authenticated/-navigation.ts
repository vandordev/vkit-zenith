import type { FileRouteTypes } from "../../routeTree.gen";

export type NavTo = FileRouteTypes["to"];
export type NavIcon =
	| "home"
	| "organizations"
	| "users"
	| "staff"
	| "activity"
	| "sessions"
	| "overview"
	| "members"
	| "api-keys"
	| "jobs";
export type NavItem = {
	id: string;
	label: string;
	to: NavTo;
	params?: Record<string, string>;
	search?: { returnTo?: string };
	icon: NavIcon;
	active: boolean;
	dividerAfter?: boolean;
};
export type NavContext = { kind: "organization" | "user" | "staff"; id: string; title: string };
export type SidebarContextData = {
	title: string;
	back: { label: string; to: NavTo; params?: Record<string, string>; search?: Record<string, unknown> };
};
export type ResolvedNavigation = {
	panel: { key: string; depth: 0 | 1 };
	items: NavItem[];
	sidebarContext?: SidebarContextData;
};

const organizationPath = /^\/organizations\/([^/]+)(?:\/|$)/;
const userPath = /^\/users\/([^/]+)(?:\/|$)/;
const staffPath = /^\/staff\/([^/]+)(?:\/|$)/;
const under = (pathname: string, base: string) => pathname === base || pathname.startsWith(`${base}/`);
function userBack(returnTo?: string): SidebarContextData["back"] {
	if (!returnTo) return { label: "Semua pengguna", to: "/users" };
	const url = new URL(returnTo, "https://bo.invalid");
	const search: Record<string, unknown> = Object.fromEntries(url.searchParams);
	if (search.page) search.page = Number(search.page);
	if (url.pathname === "/users") return { label: "Kembali ke daftar asal", to: "/users", search };
	const members = /^\/organizations\/([0-9a-f-]+)\/members$/.exec(url.pathname);
	if (members?.[1])
		return { label: "Kembali ke daftar asal", to: "/organizations/$organizationId/members", params: { organizationId: members[1] }, search };
	return { label: "Semua pengguna", to: "/users" };
}

// Pure: no React, router or data access.
export function resolveNavigation(input: {
	pathname: string;
	search: { returnTo?: string; customerReturnTo?: string };
	role: "superadmin" | "admin";
	context?: NavContext;
}): ResolvedNavigation {
	const { pathname, search, role, context } = input;
	const organizationId = organizationPath.exec(pathname)?.[1];
	if (organizationId) {
		const base = `/organizations/${organizationId}`;
		const link = { params: { organizationId } };
		const item = (id: NavIcon, label: string, to: NavTo, segment: string): NavItem => ({
			id,
			label,
			to,
			icon: id,
			active: segment ? under(pathname, `${base}/${segment}`) : pathname === base || pathname === `${base}/`,
			...link,
		});
		return {
			panel: { key: "organization", depth: 1 },
			items: [
				item("overview", "Ringkasan", "/organizations/$organizationId", ""),
				item("members", "Anggota", "/organizations/$organizationId/members", "members"),
				item("api-keys", "API Key", "/organizations/$organizationId/api-keys", "api-keys"),
				item("activity", "Log Aktivitas", "/organizations/$organizationId/activity", "activity"),
			],
			sidebarContext: {
				title: context?.kind === "organization" && context.id === organizationId ? context.title : "Organisasi",
				back: { label: "Semua organisasi", to: "/organizations" },
			},
		};
	}
	const userId = userPath.exec(pathname)?.[1];
	if (userId) {
		const base = `/users/${userId}`;
		const returnLink = search.customerReturnTo ? { search: { returnTo: search.customerReturnTo } } : {};
		const item = (id: NavIcon, label: string, to: NavTo, segment: string): NavItem => ({
			id,
			label,
			to,
			icon: id,
			active: segment ? under(pathname, `${base}/${segment}`) : pathname === base || pathname === `${base}/`,
			params: { userId },
			...returnLink,
		});
		return {
			panel: { key: "user", depth: 1 },
			items: [
				item("overview", "Ringkasan", "/users/$userId", ""),
				item("organizations", "Organisasi", "/users/$userId/organizations", "organizations"),
				item("sessions", "Sesi login", "/users/$userId/sessions", "sessions"),
				item("activity", "Log Aktivitas", "/users/$userId/activity", "activity"),
			],
			sidebarContext: {
				title: context?.kind === "user" && context.id === userId ? context.title : "Pengguna",
				back: userBack(search.customerReturnTo),
			},
		};
	}
	// Staff context is superadmin-only; an admin stays on the global panel (the route itself denies).
	const staffId = role === "superadmin" ? staffPath.exec(pathname)?.[1] : undefined;
	if (staffId) {
		const base = `/staff/${staffId}`;
		const returnSearch = search.returnTo ? new URL(search.returnTo, "https://bo.invalid").searchParams : undefined;
		const returnLink = search.returnTo ? { search: { returnTo: search.returnTo } } : {};
		const item = (id: NavIcon, label: string, to: NavTo, segment: string): NavItem => ({
			id,
			label,
			to,
			icon: id,
			active: segment ? under(pathname, `${base}/${segment}`) : pathname === base || pathname === `${base}/`,
			params: { staffId },
			...returnLink,
		});
		return {
			panel: { key: "staff", depth: 1 },
			items: [
				item("overview", "Ringkasan", "/staff/$staffId", ""),
				item("sessions", "Sesi", "/staff/$staffId/sessions", "sessions"),
				item("activity", "Log Aktivitas", "/staff/$staffId/activity", "activity"),
			],
			sidebarContext: {
				title: context?.kind === "staff" && context.id === staffId ? context.title : "Staf",
				back: {
					label: "Semua staf",
					to: "/staff",
					...(returnSearch
						? {
							search: {
								q: returnSearch.get("q") ?? "",
								status: returnSearch.get("status") ?? "",
								page: Number(returnSearch.get("page") ?? 1),
							},
						}
						: {}),
				},
			},
		};
	}
	const global = (id: NavIcon, label: string, to: NavTo, dividerAfter = false): NavItem => ({
		id,
		label,
		to,
		icon: id,
		active: to === "/" ? pathname === "/" : under(pathname, to),
		dividerAfter,
	});
	const superadmin = role === "superadmin";
	return {
		panel: { key: "global", depth: 0 },
		items: [
			global("home", "Beranda", "/"),
			global("organizations", "Organisasi", "/organizations"),
			global("users", "Pengguna", "/users", true),
			// Operasional; the divider before Sistem exists only when Sistem does.
			global("jobs", "Job", "/operations/jobs", superadmin),
			...(superadmin ? [global("staff", "Staf", "/staff"), global("activity", "Log Aktivitas", "/activity")] : []),
		],
	};
}

