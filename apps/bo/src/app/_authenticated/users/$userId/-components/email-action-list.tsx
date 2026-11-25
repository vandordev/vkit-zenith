import { DateTime } from "@repo/ui/date-time";
import { Collection } from "@repo/ui/collection";
import type { V1OrganizationListPagination, V1UserEmailAction } from "@bo/lib/api/generated/types.gen";
import { ContentState } from "../../../../-components/content-state";
import { listState, type SectionState } from "../../../../-components/org-section";
import { Pager } from "../../../../-components/pager";
import { emailActionStatus } from "../../../../-components/session-status";
import { StatusBadge } from "../../../../-components/status-badge";

const purposes: Record<string, string> = {
	verify_email: "Verifikasi email",
	setup_password: "Atur kata sandi",
	reset_password: "Atur ulang kata sandi",
};
const when = (value: string | null) => (value ? <DateTime value={value} /> : null);

/** Email-action metadata of one user: purpose and timestamps only, never the token or its hash. */
export function EmailActionList({
	actions,
	state,
	error,
	pagination,
	loading,
	now = Date.now(),
	onPage,
	onRetry,
}: {
	actions: readonly V1UserEmailAction[];
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	/** Epoch ms used to tell expired from pending; stories pin it. */
	now?: number;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	const shown = listState(state, actions.length, pagination?.total);
	return (
		<section className="min-w-0 space-y-4 border-t border-border pt-5" aria-label="Tindakan email">
			<h2 className="text-lg font-semibold">Tindakan email</h2>
			<p className="max-w-[65ch] text-xs text-muted-foreground">
				Hanya jenis dan waktu yang ditampilkan. Token tidak pernah dapat dilihat.
			</p>
			{shown === "pending" && <ContentState kind="loading" title="Memuat tindakan email…" />}
			{shown === "error" && <ContentState kind="error" title={error ?? "Tindakan email tidak dapat dimuat."} onRetry={onRetry} />}
			{shown === "empty" && (
				<ContentState
					kind="empty"
					title="Belum ada tindakan email"
					description="Verifikasi dan pengaturan kata sandi akan tercatat di sini."
				/>
			)}
			{shown === "ready" && (
				<>
					<Collection
						label="Tindakan email pelanggan"
						items={actions}
						getKey={(action) => action.id}
						columns={[
							{
								id: "purpose",
								header: "Jenis",
								cell: (action) => <span className="font-medium">{purposes[action.purpose] ?? action.purpose}</span>,
							},
							{
								id: "status",
								header: "Status",
								cell: (action) => <StatusBadge kind="emailAction" status={emailActionStatus(action, now)} />,
							},
							{ id: "created", header: "Dibuat", inlineLabel: true, cell: (action) => when(action.createdAt) },
							{ id: "expires", header: "Kedaluwarsa", inlineLabel: true, cell: (action) => when(action.expiresAt) },
							{ id: "consumed", header: "Digunakan", inlineLabel: true, cell: (action) => when(action.consumedAt) },
							{ id: "revoked", header: "Dicabut", inlineLabel: true, cell: (action) => when(action.revokedAt) },
						]}
					/>
					{pagination && <Pager pagination={pagination} noun="tindakan" loading={loading} onPage={onPage} />}
				</>
			)}
		</section>
	);
}
