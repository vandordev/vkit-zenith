import type { ReactElement, ReactNode } from "react";
import { Link } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyTitle } from "@repo/ui/empty";
import {
	ErrorState,
	ErrorStateActions,
	ErrorStateContent,
	ErrorStateDescription,
	ErrorStateHeader,
	ErrorStateTitle,
} from "@repo/ui/error-state";
import { Loading } from "@repo/ui/loading";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { Tabs } from "@repo/ui/tabs";
import { Collection, type CollectionColumn, type CollectionLayout } from "@repo/ui/collection";
import type { V1Page } from "../../../../lib/api/generated/types.gen";
import { linkProps } from "./navigation";

export type ListPageTab = {
	key: string;
	label: string;
	to: string;
	active: boolean;
	search?: Record<string, unknown>;
	params: Record<string, string>;
};
export const simulatedLabel = "Simulasi (data contoh)";

/** Banner shown while the page displays a dummy write result instead of refetched data. */
export function SimulatedBanner() {
	return (
		<p role="status" className="rounded-md border border-info/30 bg-info/10 px-3 py-2 text-sm text-info-text">
			{simulatedLabel}: hasil tindakan ini belum tersimpan dan hilang saat Anda berpindah halaman.
		</p>
	);
}

/** Server-side page controls; the total is an exact decimal string, so the maths is BigInt. */
export function Pager({
	pagination,
	noun,
	loading,
	onPage,
}: {
	pagination: V1Page;
	noun: string;
	loading?: boolean;
	onPage: (page: number) => void;
}) {
	const total = BigInt(pagination.total);
	const size = BigInt(pagination.size);
	const pages = total === 0n ? 1n : (total + size - 1n) / size;
	if (pages === 1n && pagination.page === 1) return null;
	return (
		<nav aria-label="Halaman" className="flex flex-wrap items-center justify-between gap-3 text-sm">
			<p role="status" className="text-muted-foreground">
				Halaman {pagination.page} dari {String(pages)} · {pagination.total} {noun}
			</p>
			<div className="flex gap-2">
				<Button
					type="button"
					variant="outline"
					disabled={loading || pagination.page <= 1}
					onClick={() => onPage(Math.max(1, Math.min(pagination.page - 1, Number(pages))))}
				>
					Sebelumnya
				</Button>
				<Button
					type="button"
					variant="outline"
					disabled={loading || BigInt(pagination.page) >= pages}
					onClick={() => onPage(pagination.page + 1)}
				>
					Berikutnya
				</Button>
			</div>
		</nav>
	);
}

/** The one list pattern: header + primary action → tabs → filters → rows → pager, with empty/error/pending. */
export function ListPage<Item>({
	title,
	description,
	primaryAction,
	tabs,
	filters,
	items,
	columns,
	getKey,
	getLink,
	layout,
	pagination,
	noun,
	state,
	filtered,
	emptyTitle,
	emptyDescription,
	emptyAction,
	errorMessage,
	simulated,
	loading,
	after,
	onPage,
	onRetry,
}: {
	title: string;
	description?: string;
	primaryAction?: ReactNode;
	tabs?: readonly ListPageTab[];
	filters?: ReactNode;
	items: readonly Item[];
	columns: readonly CollectionColumn<Item>[];
	getKey: (item: Item) => string;
	/** The record's page (a router `Link` without children); the whole row opens it. */
	getLink?: (item: Item) => ReactElement | undefined;
	/** `cards` for records with a face (an avatar); everything else keeps the default table. */
	layout?: CollectionLayout;
	pagination?: V1Page;
	noun: string;
	state: "ready" | "pending" | "error";
	/** A filter is active: the empty state says "no match" instead of "none yet". */
	filtered?: boolean;
	emptyTitle: string;
	emptyDescription: string;
	emptyAction?: ReactNode;
	errorMessage?: string;
	simulated?: boolean;
	loading?: boolean;
	/** A further section of the same page (Provider & Routing's per-method routes), below the list. */
	after?: ReactNode;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	// A bookmarked or shrunken page past the data (total may even be 0 now): no rows on page > 1. Keep the pager so the user can step back to page 1.
	const outOfRange = state === "ready" && items.length === 0 && pagination !== undefined && pagination.page > 1;
	const empty = state === "ready" && items.length === 0 && !outOfRange;
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader title={title} description={description} actions={primaryAction} />
			{tabs && (
				<Tabs
					variant="underline"
					aria-label={`Bagian ${title}`}
					items={tabs.map((tab) => ({
						key: tab.key,
						active: tab.active,
						link: (
							<Link {...linkProps(tab)} activeOptions={{ exact: true }}>
								{tab.label}
							</Link>
						),
					}))}
				/>
			)}
			{simulated && <SimulatedBanner />}
			{filters}
			{state === "pending" && (
				<div className="flex min-h-32 items-center justify-center py-8" aria-busy="true">
					<Loading variant="skeleton" size={48} aria-label={`Memuat ${title.toLowerCase()}`} />
				</div>
			)}
			{state === "error" && (
				<ErrorState role="alert">
					<ErrorStateContent>
						<ErrorStateHeader>
							<ErrorStateTitle>Data tidak dapat dimuat</ErrorStateTitle>
							{errorMessage && <ErrorStateDescription>{errorMessage}</ErrorStateDescription>}
						</ErrorStateHeader>
						<ErrorStateActions>
							<Button type="button" variant="outline" onClick={onRetry}>
								Muat ulang
							</Button>
						</ErrorStateActions>
					</ErrorStateContent>
				</ErrorState>
			)}
			{empty && (
				<Empty>
					<EmptyHeader>
						<EmptyTitle>{filtered ? `Tidak ada ${noun} yang cocok` : emptyTitle}</EmptyTitle>
						<EmptyDescription>{filtered ? "Ubah atau kosongkan filter untuk mencari kembali." : emptyDescription}</EmptyDescription>
					</EmptyHeader>
					{!filtered && emptyAction && <EmptyContent>{emptyAction}</EmptyContent>}
				</Empty>
			)}
			{outOfRange && pagination && (
				<>
					<Empty>
						<EmptyHeader>
							<EmptyTitle>{`Tidak ada ${noun} di halaman ini`}</EmptyTitle>
							<EmptyDescription>
								{pagination.total === "0"
									? `Tidak ada ${noun} tersisa. Gunakan tombol halaman untuk kembali ke halaman pertama.`
									: "Halaman ini melewati data yang tersedia. Gunakan tombol halaman untuk kembali."}
							</EmptyDescription>
						</EmptyHeader>
					</Empty>
					<Pager pagination={pagination} noun={noun} loading={loading} onPage={onPage} />
				</>
			)}
			{state === "ready" && !empty && !outOfRange && (
				<>
					<Collection label={title} items={items} columns={columns} getKey={getKey} getLink={getLink} layout={layout} />
					{pagination && <Pager pagination={pagination} noun={noun} loading={loading} onPage={onPage} />}
				</>
			)}
			{after}
		</PageContainer>
	);
}
