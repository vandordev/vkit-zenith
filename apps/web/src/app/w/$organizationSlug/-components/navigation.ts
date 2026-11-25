import type { LinkProps } from "@tanstack/react-router";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";

export type OrgContext = { organization: V1MyOrganization; can: (permission: string) => boolean };
export type NavGroup = "organization" | "developers";
export type OrgNavItem = {
	id: string;
	label: string;
	to: string;
	params: { organizationSlug: string };
	active: boolean;
	group?: NavGroup;
};
export type OrgNavigation = { panel: "main" | "settings"; items: OrgNavItem[] };

export const groupLabels: Record<NavGroup, string> = {
	organization: "Organisasi",
	developers: "Pengembang",
};

const root = "/w/$organizationSlug";
const under = (pathname: string, base: string) => pathname === base || pathname.startsWith(`${base}/`);

// Pure: no React, router or data access. Registers only routes that exist; features append theirs.
export function resolveOrgNavigation(input: { pathname: string; slug: string }): OrgNavigation {
	const { pathname, slug } = input;
	const base = `/w/${slug}`;
	const item = (id: string, label: string, to: string, active: boolean, group?: NavGroup): OrgNavItem => ({
		id,
		label,
		to,
		active,
		params: { organizationSlug: slug },
		...(group && { group }),
	});
	if (under(pathname, `${base}/settings`))
		return {
			panel: "settings",
			items: [
				item("profile", "Profil & Branding", `${root}/settings/profile`, under(pathname, `${base}/settings/profile`), "organization"),
				item("members", "Anggota", `${root}/settings/members`, under(pathname, `${base}/settings/members`), "organization"),
				item("api-keys", "API Key", `${root}/settings/api-keys`, under(pathname, `${base}/settings/api-keys`), "developers"),
			],
		};
	return {
		panel: "main",
		items: [
			item("overview", "Ringkasan", root, pathname === base || pathname === `${base}/`),
			item("settings", "Pengaturan", `${root}/settings/profile`, false),
		],
	};
}

export type CommandItem = { id: string; label: string; group: string; link: RouteLink };

/** Command palette destinations: every registered page in both panels, then the other organizations. */
export function resolveCommandItems(input: { slug: string; organizations: readonly { slug: string; name: string }[] }): CommandItem[] {
	const { slug } = input;
	const panel = (pathname: string) => resolveOrgNavigation({ pathname, slug }).items;
	const page = (group: string) => (item: OrgNavItem) => ({ id: item.id, label: item.label, group, link: item });
	return [
		// The main panel's Pengaturan item only opens Profil, which the settings group already lists.
		...panel(`/w/${slug}`)
			.filter((item) => item.id !== "settings")
			.map(page("Halaman")),
		...panel(`/w/${slug}/settings`).map(page("Pengaturan")),
		...input.organizations
			.filter((organization) => organization.slug !== slug)
			.map((organization) => ({
				id: `organization-${organization.slug}`,
				label: organization.name,
				group: "Ganti organisasi",
				link: orgLink(organization.slug, ""),
			})),
	];
}

/** A link inside the organization with any extra params/search. */
export function orgLink(
	slug: string,
	segment: string,
	params: Record<string, string> = {},
	search: Record<string, unknown> = {},
): RouteLink {
	return { to: `/w/$organizationSlug${segment}`, params: { organizationSlug: slug, ...params }, search };
}

export type ListTab = {
	key: string;
	label: string;
	to: string;
	params: Record<string, string>;
	search?: Record<string, unknown>;
	active: boolean;
};
export type RouteLink = { to: string; params?: Record<string, string>; search?: Record<string, unknown> };
/** `to` is a route-pattern string (typed routes only exist for registered menus); this is the single typed escape for Link props. */
export const linkProps = (link: RouteLink) => link as unknown as LinkProps;
