import type { ReactElement } from "react";
import { Button } from "@repo/ui/button";
import { Collection } from "@repo/ui/collection";
import { DateTime } from "@repo/ui/date-time";
import { FilterBar } from "@repo/ui/filter-bar";
import { CollectionSummary } from "../../../-components/collection-summary";
import { labels } from "../../../-components/common";
import { ContentState } from "../../../-components/content-state";
import { filterLabels } from "../../../-components/filter-labels";
import { Pager } from "../../../-components/pager";
import { useSearchDraft } from "../../../-components/search-draft";
import type { BoV1ListStaffResponse } from "../../../../lib/api/generated/types.gen";

type Staff = BoV1ListStaffResponse["items"][number];

export function StaffCollection({
	data,
	q,
	status,
	loading,
	error,
	onFilter,
	onPage,
	onReload,
	onOpen,
	onCreate,
	renderLink,
	navigationKey = "",
}: {
	data?: BoV1ListStaffResponse;
	q: string;
	status: string;
	loading?: boolean;
	error?: string;
	onFilter: (q: string, status: string) => void;
	onPage: (page: number) => void;
	onReload: () => void;
	/** Opens a staff record when there is no `renderLink` (previews); the route supplies a real link instead. */
	onOpen: (id: string) => void;
	onCreate?: () => void;
	/** The staff page, a router `Link` without children; the whole row opens it. */
	renderLink?: (item: Staff) => ReactElement;
	navigationKey?: string;
}) {
	const search = useSearchDraft(q, (next) => onFilter(next, status), navigationKey);
	const items = data?.items ?? [];
	const filtered = Boolean(q || status);
	return (
		<>
			<CollectionSummary
				ariaLabel="Ringkasan staf"
				metrics={
					data && [
						{ key: "total", label: "Total staf", value: data.summary.total },
						{ key: "active", label: "Aktif", value: data.summary.active },
						{ key: "disabled", label: "Dinonaktifkan", value: data.summary.disabled },
						{ key: "superadmin", label: "Superadmin", value: data.summary.superadmin },
						{ key: "admin", label: "Admin", value: data.summary.admin },
					]
				}
			/>
			<FilterBar
				search={{ value: search.draft, onChange: search.change, placeholder: "Cari nama atau email…", label: "Cari staf" }}
				filters={[
					{
						id: "status",
						label: "Status",
						options: [
							{ value: "active", label: "Aktif" },
							{ value: "disabled", label: "Dinonaktifkan" },
						],
						value: status,
						onChange: (value) => onFilter(search.settle(), value),
					},
				]}
				onClear={() => onFilter(search.settle(""), "")}
				labels={filterLabels}
			/>
			{error && !items.length ? (
				<ContentState kind="error" title={error} onRetry={onReload} />
			) : loading && !data ? (
				<ContentState kind="loading" title="Memuat data…" />
			) : items.length ? (
				<Collection
					label="Daftar staf"
					items={items}
					getKey={(item) => item.id}
					getLink={renderLink}
					columns={[
						{
							id: "name",
							header: "Staf",
							cell: (item) =>
								renderLink ? (
									item.displayName
								) : (
									<Button type="button" variant="link" className="h-auto p-0 text-left whitespace-normal" onClick={() => onOpen(item.id)}>
										{item.displayName}
									</Button>
								),
						},
						{ id: "status", header: "Status", cell: (item) => labels[item.status] },
						{ id: "email", header: "Email", cell: (item) => <span className="break-all">{item.email}</span> },
						{ id: "role", header: "Peran", inlineLabel: true, cell: (item) => labels[item.role] },
						{ id: "updated", header: "Diperbarui", inlineLabel: true, cell: (item) => <DateTime value={item.updatedAt} /> },
					]}
				/>
			) : (
				<ContentState
					kind="empty"
					title={filtered ? "Tidak ada hasil yang cocok" : "Belum ada data"}
					description={
						filtered
							? "Ubah atau kosongkan filter untuk mencari kembali."
							: onCreate
								? "Gunakan tombol tambah untuk membuat data pertama."
								: undefined
					}
					action={
						filtered ? (
							<Button type="button" variant="outline" onClick={() => onFilter(search.settle(""), "")}>
								Kosongkan filter
							</Button>
						) : (
							onCreate && <Button onClick={onCreate}>Tambah staf</Button>
						)
					}
				/>
			)}
			{items.length > 0 && loading && !error && (
				<p role="status" className="text-sm text-muted-foreground">
					Memperbarui data…
				</p>
			)}
			{data?.pagination && (
				// Staff totals are numbers; the shared pager takes the decimal-string form.
				<Pager pagination={{ ...data.pagination, total: String(data.pagination.total) }} noun="staf" loading={loading} onPage={onPage} />
			)}
		</>
	);
}
