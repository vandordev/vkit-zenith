import type { V1Actor, V1HistoryEntry } from "@bo/lib/api/generated/types.gen";

// Deterministic presentation fixtures, derived from the generated DTOs.
const none = { apiKeyPrefix: null, id: null, name: null, role: null, system: null } satisfies Omit<V1Actor, "kind">;
export const actors = {
	staff: { ...none, kind: "staff", id: "00000000-0000-4000-8000-0000000000a1", name: "Nadia Prameswari", role: "superadmin" },
	user: { ...none, kind: "user", id: "00000000-0000-4000-8000-0000000000a2", name: "Bayu Santoso", role: "owner" },
	client: { ...none, kind: "client", name: "Integrasi toko online", apiKeyPrefix: "zk_ab12" },
	system: { ...none, kind: "system", name: "Sistem", system: "worker.example_job" },
	anonymous: { ...none, kind: "anonymous", name: "Pembeli / publik" },
	unknownSystem: { ...none, kind: "system", name: "Sistem", system: "worker.future_job" },
	longName: {
		...none,
		kind: "staff",
		name: "Muhammad Alexander Wijayakusuma Hadiningrat Prawirodirjo Santoso",
		role: "admin",
	},
} as const satisfies Record<string, V1Actor>;

const entry = (n: number, actor: V1Actor, rest: Partial<V1HistoryEntry>): V1HistoryEntry => ({
	id: `00000000-0000-4000-8000-00000000b${String(n).padStart(3, "0")}`,
	occurredAt: `2026-11-01T0${n}:00:00Z`,
	actor,
	action: "organization.updated",
	outcome: "success",
	reason: null,
	targetType: "organization",
	targetId: "00000000-0000-4000-8000-0000000000c1",
	...rest,
});
export const history: V1HistoryEntry[] = [
	entry(8, actors.staff, { action: "organization.suspended", reason: "Organisasi diminta berhenti sementara oleh pemiliknya." }),
	entry(7, actors.user, { action: "membership.added", targetType: "organization_membership", outcome: "success" }),
	entry(6, actors.client, { action: "api_key.used", targetType: "api_key", outcome: "success" }),
	entry(5, actors.system, { action: "report.generated", targetType: "organization", outcome: "failure" }),
	entry(4, actors.anonymous, { action: "auth.login", targetType: "user", outcome: "denied" }),
	entry(3, actors.unknownSystem, { action: "organization.synced", reason: "Pekerjaan sistem baru dari versi API yang lebih baru." }),
	entry(2, actors.longName, {
		action: "api_key.revoked",
		targetType: "api_key",
		reason:
			"Kunci dicabut karena bocor di repositori publik; pemilik sudah menerbitkan kunci pengganti dan meminta pencabutan segera.\nDitindaklanjuti lewat tiket dukungan.",
	}),
];
