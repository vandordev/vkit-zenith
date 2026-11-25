import { DateTime } from "@repo/ui/date-time";
import { Collection } from "@repo/ui/collection";
import type { V1OrganizationListPagination, V1StaffSession } from "@bo/lib/api/generated/types.gen";
import { listState, OrgSection, type SectionState } from "../../../../../-components/org-section";
import { Pager } from "../../../../../-components/pager";
import { sessionStatus } from "../../../../../-components/session-status";
import { StatusBadge } from "../../../../../-components/status-badge";

const when = (value: string | null) => (value ? <DateTime value={value} /> : null);

/** Session metadata of one staff account; the secret, IP and device are never part of the contract. */
export function StaffSessions({
	sessions,
	state,
	error,
	pagination,
	loading,
	now = Date.now(),
	onPage,
	onRetry,
}: {
	sessions: readonly V1StaffSession[];
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	/** Epoch ms used to tell expired from active; stories pin it. */
	now?: number;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	return (
		<OrgSection
			title="Sesi staf"
			description="Metadata sesi masuk staf. Nilai sesi, alamat IP dan perangkat tidak pernah ditampilkan."
			state={listState(state, sessions.length, pagination?.total)}
			pager={pagination && <Pager pagination={pagination} noun="sesi" loading={loading} onPage={onPage} />}
			error={error}
			onRetry={onRetry}
			emptyTitle="Belum ada sesi"
			emptyDescription="Staf ini belum pernah masuk."
		>
			<Collection
				label="Daftar sesi staf"
				items={sessions}
				getKey={(session) => session.id}
				columns={[
					{ id: "id", header: "Sesi", cell: (session) => <span className="font-mono text-xs">{session.id}</span> },
					{ id: "status", header: "Status", cell: (session) => <StatusBadge kind="session" status={sessionStatus(session, now)} /> },
					{ id: "created", header: "Dibuat", inlineLabel: true, cell: (session) => when(session.createdAt) },
					{ id: "expires", header: "Kedaluwarsa", inlineLabel: true, cell: (session) => when(session.expiresAt) },
					{ id: "revoked", header: "Dicabut", inlineLabel: true, cell: (session) => when(session.revokedAt) },
				]}
			/>
		</OrgSection>
	);
}
