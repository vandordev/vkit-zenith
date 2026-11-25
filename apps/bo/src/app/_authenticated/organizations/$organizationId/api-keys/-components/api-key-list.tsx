import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { Collection } from "@repo/ui/collection";
import type { ApiKey, V1DummyResultApiKey, V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";
import { listState, OrgSection, type SectionState } from "../../../../../-components/org-section";
import { Pager } from "../../../../../-components/pager";
import { StatusBadge } from "../../../../../-components/status-badge";

export type ApiKeyListProps = {
	keys: readonly ApiKey[];
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	onPage: (page: number) => void;
	onRetry: () => void;
	onRevoke: (key: ApiKey) => void;
};

export const revokeApiKeyDialog = (key: ApiKey) => ({
	title: "Cabut API key",
	description: `${key.name} (${key.tokenPrefix}…) berhenti berlaku seketika. Tindakan ini tidak dapat dibatalkan.`,
	confirmLabel: "Cabut API key",
	tone: "destructive" as const,
});
export const revokeApiKeyResult = ({ result }: V1DummyResultApiKey) => (
	<p className="flex flex-wrap items-center gap-2 text-sm">
		{result.name} sekarang <StatusBadge kind="apiKey" status={result.status} />
	</p>
);

const when = (value: string | null) => (value ? <DateTime value={value} /> : null);

export function ApiKeyList({ keys, state, error, pagination, loading, onPage, onRetry, onRevoke }: ApiKeyListProps) {
	return (
		<OrgSection
			title="API Key"
			description="Hanya awalan kunci yang ditampilkan. Nilai kunci tidak pernah dapat dilihat."
			state={listState(state, keys.length, pagination?.total)}
			pager={pagination && <Pager pagination={pagination} noun="API key" loading={loading} onPage={onPage} />}
			error={error}
			onRetry={onRetry}
			emptyTitle="Belum ada API key"
			emptyDescription="Organisasi ini belum menerbitkan API key."
		>
			<Collection
				label="Daftar API key"
				items={keys}
				getKey={(key) => key.id}
				columns={[
					{ id: "name", header: "Nama", cell: (key) => <span className="font-medium">{key.name}</span> },
					{
						id: "prefix",
						header: "Awalan",
						inlineLabel: true,
						cell: (key) => <span className="font-mono text-xs">{key.tokenPrefix}…</span>,
					},
					{ id: "status", header: "Status", cell: (key) => <StatusBadge kind="apiKey" status={key.status} /> },
					{ id: "used", header: "Terakhir dipakai", inlineLabel: true, cell: (key) => when(key.lastUsedAt) },
					{ id: "expires", header: "Kedaluwarsa", inlineLabel: true, cell: (key) => when(key.expiresAt) },
					{ id: "revoked", header: "Dicabut", inlineLabel: true, cell: (key) => when(key.revokedAt) },
					{
						id: "actions",
						header: "Aksi",
						slot: "actions",
						cell: (key) =>
							key.status === "revoked" ? null : (
								<Button type="button" variant="outline" size="sm" aria-label={`Cabut ${key.name}`} onClick={() => onRevoke(key)}>
									Cabut
								</Button>
							),
					},
				]}
			/>
		</OrgSection>
	);
}
