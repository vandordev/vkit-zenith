import { useState, type ReactNode } from "react";
import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { CopyableId } from "@repo/ui/copyable-id";
import { JsonView } from "@repo/ui/json-view";
import type { V1HistoryEntry } from "@bo/lib/api/generated/types.gen";
import { AuditTimeline } from "../../../-components/audit-timeline";
import { dash, OrgSection, type SectionState } from "../../../-components/org-section";
import { StatusBadge, statusPresentation, type StatusKind } from "../../../-components/status-badge";

export type OpsFact = { label: string; value: ReactNode };

/** What every operations detail route hands its presentation: one record with its embedded history. */
export type OpsDetailProps<T> = {
	item?: T;
	history: readonly V1HistoryEntry[];
	state: SectionState;
	error?: string;
	onRetry: () => void;
};

/** A dummy action's returned record is kept as a local overlay (keyed by ID) over the fetched record. */
export function useOpsRecord<T extends { id: string }>(item: T | undefined, state: SectionState, history: readonly V1HistoryEntry[]) {
	const [overlay, setOverlay] = useState<T>();
	return { value: overlay && overlay.id === item?.id ? overlay : item, setOverlay, state, history };
}

export const statusMessage = (kind: StatusKind, status: string) => `Status menjadi ${statusPresentation(kind, status).label}.`;

/** Header + facts + payload/args + embedded Riwayat shared by operations detail pages. */
export function OpsDetailFrame({
	title,
	section,
	sectionLabel,
	id,
	statusKind,
	status,
	facts,
	related,
	actions,
	history,
	state,
	error,
	onRetry,
}: {
	title: string;
	section: "jobs";
	sectionLabel: string;
	id: string;
	statusKind: StatusKind;
	status?: string;
	facts: readonly OpsFact[];
	related?: ReactNode;
	actions?: ReactNode;
	history: readonly V1HistoryEntry[];
	state: SectionState;
	error?: string;
	onRetry: () => void;
}) {
	return (
		<OrgSection
			title={title}
			status={status ? <StatusBadge kind={statusKind} status={status} /> : undefined}
			actions={actions}
			breadcrumb={
				<Link to={`/operations/${section}`} search={{ q: "", page: 1, status: "", stateGroup: "" }} className="inline-flex items-center gap-1 hover:underline focus-visible:underline">
					<ArrowLeft aria-hidden="true" className="size-3.5" />
					{sectionLabel}
				</Link>
			}
			state={state}
			error={error}
			onRetry={onRetry}
			emptyTitle="Data tidak ditemukan"
			emptyDescription="Data ini tidak tersedia."
		>
			<div className="flex min-w-0 flex-col gap-8">
				<p className="text-xs text-muted-foreground">
					ID <CopyableId value={id} label={`ID ${title.toLowerCase()}`} truncate="middle" />
				</p>
				<section
					className="grid min-w-0 grid-cols-1 gap-x-8 gap-y-4 border-y border-border py-5 sm:grid-cols-2 lg:grid-cols-3"
					aria-label="Informasi utama"
				>
					{facts.map((fact) => (
						<div key={fact.label} className="min-w-0">
							<h2 className="text-xs text-muted-foreground">{fact.label}</h2>
							<div className="mt-1 break-words text-sm">{fact.value ?? dash}</div>
						</div>
					))}
				</section>
				{related}
				<section className="flex min-w-0 flex-col gap-3" aria-labelledby="ops-history">
					<h2 id="ops-history" className="font-semibold">
						Riwayat
					</h2>
					<AuditTimeline entries={history} state={state === "ready" ? "ready" : state} onRetry={onRetry} />
				</section>
			</div>
		</OrgSection>
	);
}

export function JsonSection({ heading, label, value }: { heading: string; label: string; value: unknown }) {
	return (
		<section className="flex min-w-0 flex-col gap-3" aria-label={heading}>
			<h2 className="font-semibold">{heading}</h2>
			<JsonView value={value} label={label} />
		</section>
	);
}
