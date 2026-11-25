import type {
	V1OrganizationGetBody,
	V1OrganizationListBody,
	V1UserChoiceListBody,
	V1UserChoiceListItem,
} from "@bo/lib/api/generated/types.gen";
export const noop = () => {};
export const organization: V1OrganizationGetBody = {
	id: "00000000-0000-4000-8000-000000000101",
	name: "Koperasi Nusantara",
	slug: "koperasi-nusantara",
	status: "active",
	revision: "1".repeat(64),
	createdAt: "2026-10-24T08:00:00Z",
	updatedAt: "2026-10-26T09:00:00Z",
	memberCount: "18",
	activeMemberCount: "15",
	activeOwnerCount: "2",
	canChangeStatus: true,
	statusBlockedReason: "",
};
export const organizations: V1OrganizationListBody = {
	items: [
		organization,
		{
			...organization,
			id: "00000000-0000-4000-8000-000000000102",
			name: "Yayasan Pendidikan Bersama",
			slug: "pendidikan-bersama",
			status: "suspended",
			memberCount: "9",
			activeMemberCount: "7",
			activeOwnerCount: "1",
		},
	],
	pagination: { page: 1, pageSize: 25, total: "2" },
	summary: { total: "2", active: "1", suspended: "1" },
};
export const owner: V1UserChoiceListItem = {
	id: "00000000-0000-4000-8000-000000000201",
	displayName: "Nadia Prameswari",
	email: "nadia@example.test",
	status: "active",
	eligible: true,
	disabledReason: "",
};
export const disabledOwner: V1UserChoiceListItem = {
	...owner,
	id: "00000000-0000-4000-8000-000000000202",
	displayName: "Pengguna nonaktif",
	status: "disabled",
	eligible: false,
	disabledReason: "Pengguna dinonaktifkan.",
};
export const ownerChoices: V1UserChoiceListBody = {
	items: [owner, disabledOwner],
	selected: null,
	pagination: { page: 1, pageSize: 25, total: "2" },
};
