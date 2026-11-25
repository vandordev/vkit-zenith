import type { V1HistoryEntry, V1UserEmailAction, V1UserSession } from "@bo/lib/api/generated/types.gen";
import { history } from "../../../-fixtures/audit";
import { customer } from "../-fixtures";

// Deterministic presentation fixtures from the generated DTOs: metadata only, never a token, hash, IP or agent.
export const now = Date.parse("2026-11-10T00:00:00Z");
const uid = (n: number) => `00000000-0000-4000-8000-0000000007${String(n).padStart(2, "0")}`;
export const userSessions: V1UserSession[] = [
	{ id: uid(1), createdAt: "2026-11-09T08:00:00Z", authenticatedAt: "2026-11-09T08:00:00Z", lastSeenAt: "2026-11-09T23:30:00Z", expiresAt: "2026-11-23T08:00:00Z", revokedAt: null },
	{ id: uid(2), createdAt: "2026-10-01T08:00:00Z", authenticatedAt: "2026-10-01T08:00:00Z", lastSeenAt: "2026-10-12T10:00:00Z", expiresAt: "2026-10-15T08:00:00Z", revokedAt: null },
	{ id: uid(3), createdAt: "2026-11-02T08:00:00Z", authenticatedAt: "2026-11-02T08:00:00Z", lastSeenAt: "2026-11-03T09:00:00Z", expiresAt: "2026-11-16T08:00:00Z", revokedAt: "2026-11-03T09:05:00Z" },
];
export const emailActions: V1UserEmailAction[] = [
	{ id: uid(11), purpose: "verify_email", createdAt: "2026-11-09T08:00:00Z", expiresAt: "2026-11-12T08:00:00Z", consumedAt: null, revokedAt: null },
	{ id: uid(12), purpose: "setup_password", createdAt: "2026-10-01T08:00:00Z", expiresAt: "2026-10-04T08:00:00Z", consumedAt: "2026-10-01T09:00:00Z", revokedAt: null },
	{ id: uid(13), purpose: "reset_password", createdAt: "2026-11-01T08:00:00Z", expiresAt: "2026-11-02T08:00:00Z", consumedAt: null, revokedAt: "2026-11-01T08:30:00Z" },
	{ id: uid(14), purpose: "reset_password", createdAt: "2026-10-20T08:00:00Z", expiresAt: "2026-10-21T08:00:00Z", consumedAt: null, revokedAt: null },
];
export const userHistory: V1HistoryEntry[] = history.map((entry) => ({ ...entry, targetType: "user", targetId: customer.id }));
