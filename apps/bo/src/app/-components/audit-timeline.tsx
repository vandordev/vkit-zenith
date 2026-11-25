import { DateTime } from "@repo/ui/date-time";
import type { V1HistoryEntry } from "@bo/lib/api/generated/types.gen";
import { ActorChip } from "./actor-chip";
import { ContentState } from "./content-state";
import { StatusBadge } from "./status-badge";

export type AuditTimelineProps = {
	entries: readonly V1HistoryEntry[];
	state: "ready" | "pending" | "error" | "empty";
	onRetry?: () => void;
	errorTitle?: string;
};

export function AuditTimeline({ entries, state, onRetry, errorTitle = "Riwayat tidak dapat dimuat" }: AuditTimelineProps) {
	if (state === "pending") return <ContentState kind="loading" title="Memuat riwayat…" />;
	if (state === "error") return <ContentState kind="error" title={errorTitle} onRetry={onRetry} />;
	if (state === "empty" || entries.length === 0)
		return <ContentState kind="empty" title="Belum ada riwayat" description="Perubahan pada data ini akan tercatat di sini." />;
	return (
		<ol aria-label="Riwayat perubahan" className="flex flex-col divide-y divide-border border-y border-border">
			{entries.map((entry) => (
				<li key={entry.id} className="flex min-w-0 flex-col gap-2 py-4">
					<div className="flex min-w-0 flex-wrap items-center justify-between gap-x-4 gap-y-1">
						<span className="min-w-0 font-mono text-xs wrap-anywhere">{entry.action}</span>
						<span className="flex shrink-0 items-center gap-2 text-xs text-muted-foreground">
							<StatusBadge kind="audit" status={entry.outcome} />
							<DateTime value={entry.occurredAt} />
						</span>
					</div>
					<ActorChip actor={entry.actor} />
					{entry.targetType && (
						<p className="text-xs text-muted-foreground wrap-anywhere">
							Target: <span className="font-mono">{entry.targetType}</span>
							{entry.targetId && <span className="font-mono"> · {entry.targetId}</span>}
						</p>
					)}
					{entry.reason && <p className="max-w-[65ch] text-sm leading-relaxed whitespace-pre-line wrap-anywhere">Alasan: {entry.reason}</p>}
				</li>
			))}
		</ol>
	);
}
