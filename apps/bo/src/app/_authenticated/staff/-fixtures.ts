import type { BoV1GetStaffResponse, BoV1ListStaffResponse, V1HistoryEntry, V1StaffSession } from "../../../lib/api/generated/types.gen";
import { history } from "../../-fixtures/audit";
import { id, other, revision, time } from "../../-fixtures";
export const governanceStaff: BoV1GetStaffResponse = {
	id,
	email: "nadia@example.test",
	displayName: "Nadia Prameswari",
	role: "superadmin",
	status: "active",
	updatedAt: time,
	revision,
};
export const staffList: BoV1ListStaffResponse = {
	items: [
		governanceStaff,
		{
			...governanceStaff,
			id: other,
			email: "staf-dengan-alamat-panjang-untuk-pemeriksaan@example.test",
			displayName: "Admin operasional seluruh cabang",
			role: "admin",
			status: "disabled",
		},
	],
	summary: { total: "42", active: "30", disabled: "12", superadmin: "2", admin: "40" },
	pagination: { page: 1, pageSize: 25, total: 2 },
};
// Staff sessions (T10): metadata only, never the session secret or its hash.
export const staffSessions: V1StaffSession[] = [
	{ id: "00000000-0000-4000-8000-000000000801", createdAt: "2026-11-09T08:00:00Z", expiresAt: "2026-11-10T08:00:00Z", revokedAt: null },
	{ id: "00000000-0000-4000-8000-000000000802", createdAt: "2026-10-01T08:00:00Z", expiresAt: "2026-10-02T08:00:00Z", revokedAt: null },
	{ id: "00000000-0000-4000-8000-000000000803", createdAt: "2026-11-02T08:00:00Z", expiresAt: "2026-11-03T08:00:00Z", revokedAt: "2026-11-02T10:00:00Z" },
];
export const staffHistory: V1HistoryEntry[] = history.map((entry) => ({ ...entry, targetType: "staff", targetId: id }));
