import { history } from "../../../-fixtures/audit";
import type { ApiKey, V1HistoryEntry, V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";

// Deterministic presentation fixtures derived from the generated DTOs. No secrets.
const uid = (n: number) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
export const organizationId = uid(101);
export const longName = "Nama yang sangat panjang sekali untuk menguji pemotongan teks 64ch";
const at = (day: number) => `2026-11-${String(day).padStart(2, "0")}T08:00:00Z`;
export const apiKeys: ApiKey[] = [
	{
		id: uid(301),
		name: "Server produksi",
		tokenPrefix: "zk_ab12",
		status: "active",
		lastUsedAt: at(10),
		expiresAt: null,
		revokedAt: null,
	},
	{
		id: uid(302),
		name: "Integrasi lama",
		tokenPrefix: "zk_cd34",
		status: "expired",
		lastUsedAt: at(1),
		expiresAt: at(5),
		revokedAt: null,
	},
	{
		id: uid(303),
		name: "Kunci bocor",
		tokenPrefix: "zk_ef56",
		status: "revoked",
		lastUsedAt: at(3),
		expiresAt: null,
		revokedAt: at(4),
	},
	{
		id: uid(304),
		name: "Belum pernah dipakai",
		tokenPrefix: "zk_gh78",
		status: "active",
		lastUsedAt: null,
		expiresAt: at(28),
		revokedAt: null,
	},
	{
		id: uid(305),
		name: longName,
		tokenPrefix: "zk_ij90",
		status: "active",
		lastUsedAt: at(9),
		expiresAt: null,
		revokedAt: null,
	},
];

/** Pagination metadata for a list of `count` items on `page` (total defaults to `count`). */
export const paged = (count: number, page = 1, size = 25, total = String(count)): V1OrganizationListPagination => ({
	page,
	pageSize: size,
	total,
});

/** Riwayat targeting the organization, one entry per actor kind (incl. an unknown subsystem). */
export const organizationHistory: V1HistoryEntry[] = history.map((entry) => ({
	...entry,
	targetType: "organization",
	targetId: organizationId,
}));
