import type { Job } from "@bo/lib/api/generated/types.gen";
import { DateTime } from "@repo/ui/date-time";
import { StatusBadge } from "../../../../-components/status-badge";
import { statusOptions } from "../../../-components/filter-options";
import { GlobalCollection, type GlobalCollectionProps } from "../../../-components/global-list";
import { jobLink, operationId } from "../../-components/ops-links";

export const jobStatuses = ["available", "scheduled", "running", "retryable", "completed", "cancelled", "discarded"] as const;
export type JobCollectionProps = GlobalCollectionProps & { items: readonly Job[]; status: string; stateGroup: string };

/** Jobs belong to no organization, so the organization filter and column do not exist here. */
export function JobCollection({ items, status, stateGroup, ...props }: JobCollectionProps) {
	return (
		<GlobalCollection
			{...props}
			title="Job"
			description="Job latar belakang beserta antrean, status dan percobaan."
			noun="job"
			searchLabel="Cari job"
			selects={[
				{ key: "status", label: "Status", value: status, options: statusOptions(jobStatuses, "job") },
				{
					key: "stateGroup",
					label: "Kelompok",
					value: stateGroup,
					options: [
						{ value: "", label: "Semua kelompok" },
						{ value: "failed", label: "Gagal atau akan diulang" },
					],
				},
			]}
			items={items}
			getKey={(item) => item.id}
			getLink={(item) => jobLink(item.id)}
			columns={[
				{ id: "id", header: "Job", cell: (item) => operationId(item.id) },
				{
					id: "kind",
					header: "Jenis",
					inlineLabel: true,
					cell: (item) => <span className="font-mono text-xs break-all">{item.kind}</span>,
				},
				{ id: "queue", header: "Antrean", inlineLabel: true, cell: (item) => item.queue },
				{ id: "status", header: "Status", cell: (item) => <StatusBadge kind="job" status={item.state} /> },
				{ id: "attempts", header: "Percobaan", align: "end", inlineLabel: true, cell: (item) => `${item.attempt}/${item.maxAttempts}` },
				{ id: "created", header: "Dibuat", cell: (item) => <DateTime value={item.createdAt} /> },
			]}
		/>
	);
}
