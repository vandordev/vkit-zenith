import type { V1HistoryEntry, V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";
import { AuditTimeline } from "../../../../../-components/audit-timeline";
import { OrgSection, type SectionState } from "../../../../../-components/org-section";
import { Pager } from "../../../../../-components/pager";

export function OrganizationActivity({
	entries,
	state,
	error,
	pagination,
	loading,
	onPage,
	onRetry,
}: {
	entries: readonly V1HistoryEntry[];
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	return (
		<OrgSection title="Log Aktivitas" description="Siapa atau apa yang mengubah data organisasi ini, beserta alasannya." state="ready">
			<AuditTimeline
				entries={entries}
				state={state === "ready" && pagination?.total === "0" ? "empty" : state}
				onRetry={onRetry}
				errorTitle={error}
			/>
			{state === "ready" && pagination && <Pager pagination={pagination} noun="catatan" loading={loading} onPage={onPage} />}
		</OrgSection>
	);
}
