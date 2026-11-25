import type { V1MyOrganization, V1OrganizationMember, V1OrganizationProfile } from "../../../../lib/api/generated/types.gen";
import { resolveOrgNavigation, type OrgNavigation } from "./navigation";

// Deterministic, fake, non-sensitive values shaped like the generated Web API DTOs.
const roles = ["owner", "admin", "member"] as const;
export const permissionsByRole: Record<V1MyOrganization["role"], string[]> = {
	owner: ["developers.manage", "members.manage", "organization.manage", "organization.read"],
	admin: ["developers.manage", "members.manage", "organization.read"],
	member: ["organization.read"],
};
export const organization = (role: V1MyOrganization["role"] = "owner", overrides: Partial<V1MyOrganization> = {}): V1MyOrganization => ({
	id: "0f3a6c1e-0000-4000-8000-000000000001",
	name: "Kopi Senja",
	slug: "kopi-senja",
	status: "active",
	hasLogo: false,
	role,
	permissions: permissionsByRole[role],
	...overrides,
});
export const organizations: V1MyOrganization[] = [
	organization("owner"),
	organization("admin", { id: "0f3a6c1e-0000-4000-8000-000000000002", name: "Toko Buku Arunika", slug: "toko-buku-arunika" }),
	organization("member", {
		id: "0f3a6c1e-0000-4000-8000-000000000003",
		name: "Studio Tari Larasati",
		slug: "studio-tari-larasati",
		status: "suspended",
	}),
];
export const longOrganization = organization("owner", {
	name: "Perkumpulan Penggiat Kopi Nusantara Tenggara Timur dan Sekitarnya Cabang Utama Sejahtera",
	slug: "perkumpulan-penggiat-kopi-nusantara-tenggara-timur-dan-sekitarnya-cabang-utama",
});

export const navigationFor = (slug: string, pathname = `/w/${slug}`): OrgNavigation => resolveOrgNavigation({ pathname, slug });

export const profile: V1OrganizationProfile = {
	id: organization().id,
	name: "Kopi Senja",
	slug: "kopi-senja",
	status: "active",
	hasLogo: false,
};

export const members: V1OrganizationMember[] = [
	{ name: "Rani Wulandari", email: "rani@example.com", role: "owner", joinedAt: "2026-01-10T03:00:00Z" },
	{ name: "Agus Prasetyo", email: "agus@example.com", role: "admin", joinedAt: "2026-02-14T03:00:00Z" },
	...Array.from({ length: 8 }, (_, index) => ({
		name: `Anggota Tim ${index + 1}`,
		email: `anggota${index + 1}@example.com`,
		role: roles[2],
		joinedAt: `2026-03-${String(index + 10).padStart(2, "0")}T03:00:00Z`,
	})),
];
export const singleOwner: V1OrganizationMember[] = members.slice(0, 1);
