import type { Job, V1HistoryEntry, V1HomeCountersBody } from "@bo/lib/api/generated/types.gen";
import { organizationHistory } from "./organizations/$organizationId/-setup-fixtures";

// Deterministic fixtures for the Beranda and operations stories. No secrets.
const uid = (n: number) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const time = "2026-11-01T08:00:00Z";
export const jobStates = ["available", "scheduled", "running", "retryable", "completed", "cancelled", "discarded"] as const;
export const jobs: Job[] = jobStates.map((state, index) => ({
	id: String(index + 1),
	args: { resourceId: uid(10 + index), options: { dryRun: false } },
	attempt: String(state === "retryable" ? 3 : state === "discarded" ? 10 : index % 3),
	createdAt: time,
	errors: state === "retryable" || state === "discarded" ? ["upstream timeout setelah 30 detik", "koneksi ditolak"] : [],
	kind: index === 6 ? "report.generate.monthly.summary.with.a.very.long.job.kind.name" : ["email.send", "report.generate", "email.send", "cleanup.sessions", "report.generate", "email.send", "cleanup.sessions"][index]!,
	maxAttempts: "10",
	queue: "default",
	state,
}));
export const opsHistory = (targetType: string, targetId: string): V1HistoryEntry[] => organizationHistory.slice(0, 3).map((entry) => ({ ...entry, targetType, targetId }));
export const counters = (count: string): V1HomeCountersBody => ({
	failedJobs: { count, href: "/operations/jobs?stateGroup=failed" },
});

/** Props every global collection story shares; stories override only what differs. */
export const collectionArgs = {
	q: "",
	state: "ready" as const,
	onFilter: () => {},
	onPage: () => {},
	onRetry: () => {},
	onFilterOrganization: () => {},
};
export const pageOf = (count: number, page = 1, total = String(count)) => ({ page, pageSize: 25, total });
/** A dummy backend answer arriving after `ms`; stories count calls around it to expose double submits. */
export const delayed = <T>(value: T, ms = 600) => new Promise<T>((resolve) => setTimeout(() => resolve(value), ms));
