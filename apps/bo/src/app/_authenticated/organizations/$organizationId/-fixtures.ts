import type { V1OrganizationMembershipGetBody, V1OrganizationMembershipListBody } from "@bo/lib/api/generated/types.gen";
import { organization } from "../-fixtures";
export const membership: V1OrganizationMembershipGetBody = {
	id: "00000000-0000-4000-8000-000000000301",
	organizationId: organization.id,
	userId: "00000000-0000-4000-8000-000000000201",
	organizationName: organization.name,
	organizationSlug: organization.slug,
	organizationStatus: "active",
	userName: "Nadia Prameswari",
	userEmail: "nadia@example.test",
	userStatus: "active",
	status: "active",
	role: "owner",
	revision: "1".repeat(64),
	createdAt: organization.createdAt,
	updatedAt: organization.updatedAt,
	canRemove: false,
	removeBlockedReason: "Cabut kepemilikan terlebih dahulu melalui tindakan terpisah.",
	canRestore: false,
	restoreBlockedReason: "Keanggotaan masih aktif.",
	canChangeRole: true,
	roleBlockedReason: "",
};
export const ordinaryMember: V1OrganizationMembershipGetBody = {
	...membership,
	id: "00000000-0000-4000-8000-000000000302",
	userId: "00000000-0000-4000-8000-000000000202",
	userName: "Bayu Santoso",
	userEmail: "bayu@example.test",
	role: "admin",
	canRemove: true,
	removeBlockedReason: "",
};
export const removedMember: V1OrganizationMembershipGetBody = {
	...ordinaryMember,
	status: "removed",
	canRemove: false,
	removeBlockedReason: "Keanggotaan sudah dihapus.",
	canRestore: true,
	restoreBlockedReason: "",
	canChangeRole: false,
	roleBlockedReason: "Kepemilikan membutuhkan keanggotaan aktif.",
};
export const soleOwner: V1OrganizationMembershipGetBody = {
	...membership,
	canChangeRole: false,
	roleBlockedReason: "Organisasi harus tetap memiliki pemilik aktif lain.",
};
export const memberships: V1OrganizationMembershipListBody = {
	items: [
		membership,
		{
			...membership,
			id: "00000000-0000-4000-8000-000000000304",
			userId: "00000000-0000-4000-8000-000000000204",
			userName: "Dewi Utami",
			userEmail: "dewi@example.test",
		},
		ordinaryMember,
		{ ...removedMember, id: "00000000-0000-4000-8000-000000000303", userName: "Citra", userEmail: "citra@example.test" },
	],
	pagination: { page: 1, pageSize: 25, total: "4" },
	summary: { total: "4", active: "3", removed: "1", owners: "2" },
};
