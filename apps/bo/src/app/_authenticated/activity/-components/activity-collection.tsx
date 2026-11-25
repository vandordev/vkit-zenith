import type { ReactElement } from "react";
import { Badge } from "@repo/ui/badge";
import { Button } from "@repo/ui/button";
import { Collection } from "@repo/ui/collection";
import { DateTime } from "@repo/ui/date-time";
import { ContentState } from "@bo/app/-components/content-state";
import { Pager } from "@bo/app/-components/pager";
import type { V1ActivityListBody, V1ActivityRow, V1ActivityAction } from "@bo/lib/api/generated/types.gen";

export const outcomeLabels = { success: "Berhasil", failure: "Gagal", denied: "Ditolak" };
export function actorLabel(actor: V1ActivityRow["actor"]) {
	return actor.kind === "anonymous" ? "Anonim" : actor.kind === "system" ? "Sistem" : (actor.name ?? actor.id ?? "Staf tidak tersedia");
}
export function targetLabel(target: V1ActivityRow["target"]) {
	return target.type === "auth_summary" ? "Ringkasan autentikasi" : (target.name ?? target.id ?? "Target tidak tersedia");
}

export function ActivityCollection({
	data,
	actions = [],
	filtered = false,
	loading,
	error,
	onPage,
	onReload,
	onReset,
	renderLink,
}: {
	data?: V1ActivityListBody;
	actions?: V1ActivityAction[];
	filtered?: boolean;
	loading?: boolean;
	error?: string;
	onPage: (page: number) => void;
	onReload: () => void;
	onReset: () => void;
	/** The activity's detail, a router `Link` without children; the whole row opens it. */
	renderLink: (row: V1ActivityRow) => ReactElement | null;
}) {
	if (!data)
		return loading || error ? (
			<ContentState kind={error ? "error" : "loading"} title={error ?? "Memuat aktivitas…"} onRetry={error ? onReload : undefined} />
		) : null;
	return (
		<div className="flex min-w-0 flex-col gap-3">
			{error && <ContentState kind="error" title={error} description="Data terakhir tetap ditampilkan." onRetry={onReload} />}
			{data.items.length ? (
				<Collection
					label="Aktivitas tercatat"
					items={data.items}
					getKey={(row) => row.id}
					getLink={(row) => renderLink(row) ?? undefined}
					columns={[
						{
							id: "action",
							header: "Tindakan",
							cell: (row) => (
								<span className="flex min-w-0 flex-col gap-0.5">
									<span>{actions.find((action) => action.value === row.action)?.label ?? row.action}</span>
									<span className="font-mono text-xs font-normal break-all text-muted-foreground">{row.action}</span>
								</span>
							),
						},
						{ id: "outcome", header: "Hasil tercatat", cell: (row) => <Badge variant="secondary">{outcomeLabels[row.outcome]}</Badge> },
						{ id: "actor", header: "Pelaku (label saat ini)", inlineLabel: true, cell: (row) => actorLabel(row.actor) },
						{ id: "target", header: "Target (label saat ini)", inlineLabel: true, cell: (row) => targetLabel(row.target) },
						{ id: "occurredAt", header: "Waktu", cell: (row) => <DateTime value={row.occurredAt} locale="id-ID" /> },
					]}
				/>
			) : (
				<ContentState
					kind="empty"
					title={filtered ? "Tidak ada aktivitas yang cocok" : "Belum ada aktivitas tercatat"}
					description={filtered ? "Ubah atau kosongkan filter untuk mencari kembali." : "Riwayat yang tersedia akan ditampilkan di sini."}
					action={
						filtered ? (
							<Button variant="outline" onClick={onReset}>
								Kosongkan filter
							</Button>
						) : undefined
					}
				/>
			)}
			{loading && (
				<p role="status" className="text-xs text-muted-foreground">
					Memperbarui aktivitas…
				</p>
			)}
			{/* Activity totals are numbers; the shared pager takes the decimal-string form. */}
			<Pager pagination={{ ...data.pagination, total: String(data.pagination.total) }} noun="aktivitas" loading={loading} onPage={onPage} />
		</div>
	);
}
