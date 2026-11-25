import type { V1UserGetBody, V1UserListBody, V1UserMembershipListBody } from "@bo/lib/api/generated/types.gen";
import { membership } from "../organizations/$organizationId/-fixtures";
export const customer: V1UserGetBody = {
	id: membership.userId,
	displayName: membership.userName,
	email: membership.userEmail,
	status: "active",
	emailVerifiedAt: null,
	revision: "1".repeat(64),
	createdAt: membership.createdAt,
	updatedAt: membership.updatedAt,
	membershipCount: "2",
	activeMembershipCount: "1",
	ownerMembershipCount: "1",
	canChangeEmail: true,
	emailBlockedReason: "",
	canVerifyEmail: true,
	verificationBlockedReason: "",
	canChangeStatus: false,
	statusBlockedReason: "Pengguna merupakan pemilik aktif terakhir di organisasi terkait.",
};
export const verifiedCustomer: V1UserGetBody = {
	...customer,
	id: "00000000-0000-4000-8000-000000000205",
	displayName: "Dewi Utami",
	email: "dewi@example.test",
	status: "disabled",
	emailVerifiedAt: "2026-10-25T08:00:00Z",
	canVerifyEmail: false,
	verificationBlockedReason: "Email sudah terverifikasi.",
	canChangeStatus: true,
	statusBlockedReason: "",
};
export const customers: V1UserListBody = {
	items: [customer, verifiedCustomer],
	pagination: { page: 1, pageSize: 25, total: "2" },
	summary: { total: "2", active: "1", disabled: "1", verified: "1", unverified: "1" },
};
export const userMemberships: V1UserMembershipListBody = {
	items: [
		{ ...membership, canChangeRole: false, roleBlockedReason: "Organisasi harus tetap memiliki pemilik aktif lain." },
		{
			...membership,
			id: "00000000-0000-4000-8000-000000000305",
			organizationId: "00000000-0000-4000-8000-000000000105",
			organizationName: "Yayasan Pendidikan Bersama",
			organizationSlug: "pendidikan-bersama",
			organizationStatus: "suspended",
			role: "member",
			status: "removed",
			canRemove: false,
			removeBlockedReason: "Keanggotaan sudah dihapus.",
			canRestore: true,
			restoreBlockedReason: "",
			canChangeRole: false,
			roleBlockedReason: "Kepemilikan membutuhkan keanggotaan aktif.",
		},
	],
	pagination: { page: 1, pageSize: 25, total: "2" },
	summary: { total: "2", active: "1", removed: "1", owners: "1" },
};
