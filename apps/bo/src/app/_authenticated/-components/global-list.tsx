import type { ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Collection, type CollectionProps } from "@repo/ui/collection";
import { CopyableId } from "@repo/ui/copyable-id";
import { FilterBar } from "@repo/ui/filter-bar";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import type { V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";
import { ContentState } from "../../-components/content-state";
import { filterLabels } from "../../-components/filter-labels";
import { useSearchDraft } from "../../-components/search-draft";
import { listState, type SectionState } from "../../-components/org-section";
import { Pager } from "../../-components/pager";

/** The record-specific half of a global list: which records and how each shows. */
export type GlobalRecords<Item> = Pick<CollectionProps<Item>, "items" | "columns" | "getKey" | "getLink">;
export type FilterSelect = { key: string; label: string; value: string; options: readonly { value: string; label: string }[] };

/** Search + selects of a global list; every change applies itself. `onFilter` restarts at page 1 in the adapter. */
export function ListFilters({
	searchLabel,
	q,
	selects,
	organizationId,
	onFilter,
	onClearOrganization,
	navigationKey,
}: {
	searchLabel: string;
	q: string;
	selects: readonly FilterSelect[];
	organizationId?: string;
	onFilter: (q: string, values: Record<string, string>) => void;
	onClearOrganization?: () => void;
	navigationKey?: string;
}) {
	const values = Object.fromEntries(selects.map((select) => [select.key, select.value]));
	const search = useSearchDraft(q, (next) => onFilter(next, values), navigationKey);
	return (
		<div className="flex min-w-0 flex-col gap-3">
			<FilterBar
				search={{ value: search.draft, onChange: search.change, placeholder: `${searchLabel}…`, label: searchLabel }}
				filters={selects.map((select) => ({
					id: select.key,
					label: select.label,
					// FilterBar supplies its own "all" choice.
					options: select.options.filter((option) => option.value !== ""),
					value: select.value,
					onChange: (value: string) => onFilter(search.settle(), { ...values, [select.key]: value }),
				}))}
				onClear={() => onFilter(search.settle(""), Object.fromEntries(selects.map((select) => [select.key, ""])))}
				labels={filterLabels}
			/>
			{organizationId && (
				<p className="flex flex-wrap items-center gap-2 text-sm">
					<span className="text-muted-foreground">Organisasi:</span>
					<CopyableId value={organizationId} label="ID organisasi" truncate="middle" />
					<Button type="button" variant="ghost" size="xs" onClick={onClearOrganization}>
						Hapus filter organisasi
					</Button>
				</p>
			)}
		</div>
	);
}

/** Shared frame of every global transaction and operations list. */
export function GlobalListPage<Item>({
	title,
	description,
	noun,
	caption,
	filters,
	filtered,
	state,
	error,
	pagination,
	loading,
	onPage,
	onRetry,
	...collection
}: GlobalRecords<Item> & {
	title: string;
	description: string;
	noun: string;
	caption: string;
	filters: ReactNode;
	filtered: boolean;
	state: SectionState;
	error?: string;
	pagination?: V1OrganizationListPagination;
	loading?: boolean;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	// A page past the data (or a total that dropped to 0 meanwhile) keeps its pager so staff can step back.
	const outOfRange = state === "ready" && collection.items.length === 0 && pagination !== undefined && pagination.page > 1;
	const shown = outOfRange ? "ready" : listState(state, collection.items.length, pagination?.total);
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader title={title} description={description} />
			{filters}
			{shown === "pending" && <ContentState kind="loading" title={`Memuat ${title.toLowerCase()}…`} />}
			{shown === "error" && <ContentState kind="error" title={error ?? "Data tidak dapat dimuat."} onRetry={onRetry} />}
			{shown === "empty" && (
				<ContentState
					kind="empty"
					title={filtered ? `Tidak ada ${noun} yang cocok` : `Belum ada ${noun}`}
					description={filtered ? "Ubah atau kosongkan filter untuk mencari kembali." : `Belum ada ${noun} yang tercatat.`}
				/>
			)}
			{shown === "ready" && outOfRange && (
				<>
					<ContentState
						kind="empty"
						title={`Tidak ada ${noun} di halaman ini`}
						description="Halaman ini melewati data yang tersedia. Gunakan tombol halaman untuk kembali."
					/>
					{pagination && <Pager pagination={pagination} noun={noun} loading={loading} onPage={onPage} />}
				</>
			)}
			{shown === "ready" && !outOfRange && (
				<>
					<Collection label={caption} {...collection} />
					{pagination && <Pager pagination={pagination} noun={noun} loading={loading} onPage={onPage} />}
				</>
			)}
		</PageContainer>
	);
}

/** Props every global collection (transactions and operations) shares with its route adapter. */
export type GlobalCollectionProps = {
	q: string;
	organizationId?: string;
	pagination?: V1OrganizationListPagination;
	state: SectionState;
	error?: string;
	loading?: boolean;
	onFilter: (q: string, values: Record<string, string>) => void;
	onClearOrganization?: () => void;
	navigationKey?: string;
	onPage: (page: number) => void;
	onRetry: () => void;
};

/** Filters + list frame; resources only supply their copy, select options and rows. */
export function GlobalCollection<Item>({
	title,
	description,
	noun,
	searchLabel,
	selects,
	items,
	columns,
	getKey,
	getLink,
	...props
}: GlobalCollectionProps &
	GlobalRecords<Item> & {
		title: string;
		description: string;
		noun: string;
		searchLabel: string;
		selects: readonly FilterSelect[];
	}) {
	return (
		<GlobalListPage
			title={title}
			description={description}
			noun={noun}
			caption={`Daftar ${noun}`}
			filtered={Boolean(props.q || props.organizationId || selects.some((select) => select.value))}
			filters={
				<ListFilters
					searchLabel={searchLabel}
					q={props.q}
					selects={selects}
					organizationId={props.organizationId}
					onFilter={props.onFilter}
					onClearOrganization={props.onClearOrganization}
					navigationKey={
						props.navigationKey ??
						JSON.stringify([props.pagination?.page, props.organizationId, selects.map(({ key, value }) => [key, value])])
					}
				/>
			}
			items={items}
			columns={columns}
			getKey={getKey}
			getLink={getLink}
			state={props.state}
			error={props.error}
			pagination={props.pagination}
			loading={props.loading}
			onPage={props.onPage}
			onRetry={props.onRetry}
		/>
	);
}
