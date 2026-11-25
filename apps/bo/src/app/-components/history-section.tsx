import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import type { V1HistoryEntry, V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";
import { AuditTimeline } from "./audit-timeline";
import { ContentState } from "./content-state";
import { Pager } from "./pager";
import { useSearchDraft } from "./search-draft";
import type { SectionState } from "./org-section";

/** Log Aktivitas of one user or staff account: searchable (action and actor), paged, applies as you type. */
export function HistorySection({
	title = "Log Aktivitas",
	description,
	q,
	entries,
	state,
	error,
	pagination,
	loading,
	onSearch,
	onPage,
	onRetry,
	navigationKey = "",
}: {
	title?: string;
	description: string;
	q: string;
	entries: readonly V1HistoryEntry[];
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	onSearch: (q: string) => void;
	onPage: (page: number) => void;
	onRetry: () => void;
	navigationKey?: string;
}) {
	const search = useSearchDraft(q, onSearch, navigationKey);
	const empty = state === "ready" && pagination?.total === "0";
	// A later page past the data keeps its pager so staff can step back.
	const outOfRange = state === "ready" && entries.length === 0 && !empty && pagination !== undefined && pagination.page > 1;
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader title={title} description={description} />
			<div className="flex flex-wrap items-end gap-3">
				<Input
					label="Cari tindakan atau pelaku"
					name="q"
					value={search.draft}
					maxLength={100}
					autoComplete="off"
					onChange={(event) => search.change(event.target.value)}
					containerClassName="min-w-0 flex-1 basis-56 sm:max-w-sm"
				/>
				{(search.draft || q) && (
					<Button type="button" variant="ghost" onClick={() => onSearch(search.settle(""))}>
						Kosongkan filter
					</Button>
				)}
			</div>
			{empty && q ? (
				<ContentState kind="empty" title="Tidak ada catatan yang cocok" description="Ubah atau kosongkan filter untuk mencari kembali." />
			) : outOfRange ? (
				<ContentState
					kind="empty"
					title="Tidak ada catatan di halaman ini"
					description="Halaman ini melewati data yang tersedia. Gunakan tombol halaman untuk kembali."
				/>
			) : (
				<AuditTimeline entries={entries} state={empty ? "empty" : state} onRetry={onRetry} errorTitle={error} />
			)}
			{state === "ready" && pagination && <Pager pagination={pagination} noun="catatan" loading={loading} onPage={onPage} />}
		</PageContainer>
	);
}
