import { useRef, type ReactNode, type RefObject } from "react";
import { Drawer, DrawerBody } from "@repo/ui/drawer";
import { DateTime } from "@repo/ui/date-time";
import { ContentState } from "@bo/app/-components/content-state";
import { labels } from "@bo/app/-components/common";
import type { V1ActivityDetail } from "@bo/lib/api/generated/types.gen";
import { actorLabel, targetLabel, outcomeLabels } from "./activity-collection";

export type ActivityDetailProps = {
	selected?: string;
	data?: V1ActivityDetail;
	loading?: boolean;
	error?: string;
	onClose: () => void;
	onRetry: () => void;
	targetLink?: ReactNode;
	returnFocus?: RefObject<HTMLElement | null>;
};
export function ActivityDetail(props: ActivityDetailProps) {
	const retained = useRef<ReactNode>(null);
	const data = props.data?.id === props.selected ? props.data : undefined;
	const content = props.error ? (
		<ContentState kind="error" title={props.error} onRetry={props.onRetry} />
	) : !data ? (
		<ContentState kind="loading" title="Memuat detail aktivitas…" />
	) : (
		<div className="space-y-6 text-sm">
			<dl className="space-y-4 break-words">
				<div>
					<dt className="text-muted-foreground">ID aktivitas</dt>
					<dd className="break-all">{data.id}</dd>
				</div>
				<div>
					<dt className="text-muted-foreground">Waktu lokal</dt>
					<dd>
						<DateTime value={data.occurredAt} locale="id-ID" />
					</dd>
				</div>
				<div>
					<dt className="text-muted-foreground">Pelaku (nama saat ini)</dt>
					<dd>{actorLabel(data.actor)}</dd>
					{data.actor.id && <dd className="break-all text-xs">{data.actor.id}</dd>}
				</div>
				<div>
					<dt className="text-muted-foreground">Tindakan tercatat</dt>
					<dd className="break-all">{data.action}</dd>
				</div>
				<div>
					<dt className="text-muted-foreground">Hasil tercatat</dt>
					<dd>{outcomeLabels[data.outcome]}</dd>
				</div>
				<div>
					<dt className="text-muted-foreground">Target (nama saat ini)</dt>
					<dd>{targetLabel(data.target)}</dd>
					{data.target.id && <dd className="break-all text-xs">{data.target.id}</dd>}
					<dd>{data.target.available ? props.targetLink : null}</dd>
				</div>
			</dl>
			<section className="space-y-3">
				<h3 className="font-semibold">Detail tercatat</h3>
				<p className="text-xs text-muted-foreground">
					Nama yang ditampilkan adalah label saat ini, bukan snapshot historis. Daftar bidang bersifat deskriptif dan bukan perbandingan
					nilai yang pasti berubah.
				</p>
				{!data.detailAvailable ? (
					<p>Detail aman tidak tersedia untuk aktivitas ini.</p>
				) : (
					<>
						<p className="break-words">Bidang tercatat: {data.changedFields.length ? data.changedFields.join(", ") : "Tidak tersedia"}</p>
						{(data.previous || data.current) && (
							<div className="space-y-2">
								<p>Nilai aman yang benar-benar dicatat:</p>
								{(["previous", "current"] as const).map((key) => (
									<p key={key}>
										{key === "previous" ? "Sebelum" : "Sesudah"}: peran {labels[data[key]?.role ?? ""] ?? "Tidak tersedia"}; status{" "}
										{labels[data[key]?.status ?? ""] ?? "Tidak tersedia"}
									</p>
								))}
							</div>
						)}
					</>
				)}
			</section>
		</div>
	);
	// Retain only exit content; a new selection never shows the previous event.
	if (props.selected) retained.current = content;
	return (
		<Drawer
			title="Detail aktivitas"
			description="Pemeriksaan read-only atas riwayat yang tersedia."
			open={!!props.selected}
			onOpenChange={(open) => {
				if (!open) props.onClose();
			}}
			onOpenChangeComplete={(open) => {
				if (!open) retained.current = null;
			}}
			contentProps={{ closeButtonLabel: "Tutup detail aktivitas", finalFocus: () => props.returnFocus?.current ?? true }}
		>
			<DrawerBody>{props.selected ? content : retained.current}</DrawerBody>
		</Drawer>
	);
}
