import type { ReactNode } from "react";
import { Badge } from "@repo/ui/badge";
import { DateTime } from "@repo/ui/date-time";
import { ContentState } from "@bo/app/-components/content-state";
import type { V1ActivityRecentBody, V1ActivityRow } from "@bo/lib/api/generated/types.gen";
import { actorLabel, targetLabel, outcomeLabels } from "../activity/-components/activity-collection";

export type RecentActivityProps = {
	data?: V1ActivityRecentBody;
	loading?: boolean;
	error?: string;
	onRetry: () => void;
	renderLink: (row: V1ActivityRow) => ReactNode;
	renderAll: () => ReactNode;
};
export function RecentActivity({ data, loading, error, onRetry, renderLink, renderAll }: RecentActivityProps) {
	return (
		<section aria-labelledby="recent-heading" className="min-w-0 space-y-4">
			<h2 id="recent-heading" className="text-base font-semibold">
				Aktivitas terbaru
			</h2>
			<p className="text-xs text-muted-foreground">
				Label pelaku dan target adalah label saat ini, bukan nama historis. Maksimal lima aktivitas tercatat.
			</p>
			{error && <ContentState kind="error" title={error} onRetry={onRetry} />}
			{loading && (
				<p role="status" className="text-sm text-muted-foreground">
					{data ? "Memperbarui aktivitas…" : "Memuat aktivitas…"}
				</p>
			)}
			{data &&
				(data.items.length ? (
					<ul className="divide-y divide-border">
						{data.items.map((row) => (
							<li key={row.id} className="space-y-2 py-3 text-sm">
								<div className="flex flex-wrap justify-between gap-2">
									<span className="break-all font-medium">{row.action}</span>
									<Badge variant="secondary">{outcomeLabels[row.outcome]}</Badge>
								</div>
								<p className="break-words">
									{actorLabel(row.actor)} · {targetLabel(row.target)}
								</p>
								<div className="flex flex-wrap items-center justify-between gap-3 text-xs">
									<DateTime value={row.occurredAt} locale="id-ID" />
									{renderLink(row)}
								</div>
							</li>
						))}
					</ul>
				) : (
					<ContentState kind="empty" title="Belum ada aktivitas tercatat" />
				))}
			{renderAll()}
		</section>
	);
}
