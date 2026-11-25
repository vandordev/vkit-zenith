import type { ReactElement } from "react";
import { Building2, CircleCheck, PauseCircle } from "lucide-react";
import { Button } from "@repo/ui/button";
import { Badge } from "@repo/ui/badge";
import { Collection } from "@repo/ui/collection";
import { DateTime } from "@repo/ui/date-time";
import { FilterBar } from "@repo/ui/filter-bar";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { StatsCard } from "@repo/ui/stats-card";
import { ContentState } from "../../../-components/content-state";
import { filterLabels } from "../../../-components/filter-labels";
import { useSearchDraft } from "../../../-components/search-draft";
import type { V1OrganizationListBody, V1OrganizationListItem } from "@bo/lib/api/generated/types.gen";

export function OrganizationCollection({
	data,
	q,
	status,
	loading,
	error,
	onCreate,
	onFilter,
	onPage,
	onRetry,
	organizationLink,
	navigationKey = "",
}: {
	data?: V1OrganizationListBody;
	q: string;
	status: "" | "active" | "suspended";
	loading?: boolean;
	error?: string;
	onCreate: () => void;
	onFilter: (q: string, status: "" | "active" | "suspended") => void;
	onPage: (page: number) => void;
	onRetry: () => void;
	/** The organization's page, a router `Link` without children; the whole row opens it. */
	organizationLink: (row: V1OrganizationListItem) => ReactElement;
	navigationKey?: string;
}) {
	const search = useSearchDraft(q, (next) => onFilter(next, status), navigationKey);
	const filtered = Boolean(q || status);
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader
				title="Organisasi"
				description="Profil organisasi, anggota dan kepemilikan pelanggan."
				actions={<Button onClick={onCreate}>Tambah organisasi</Button>}
			/>
			<StatsCard
				ariaLabel="Ringkasan seluruh organisasi"
				items={[
					{ key: "total", title: "Total organisasi", value: { display: data?.summary.total ?? "—" }, icon: Building2 },
					{ key: "active", title: "Aktif", value: { display: data?.summary.active ?? "—" }, icon: CircleCheck },
					{ key: "suspended", title: "Ditangguhkan", value: { display: data?.summary.suspended ?? "—" }, icon: PauseCircle },
				]}
			/>
			<FilterBar
				search={{ value: search.draft, onChange: search.change, placeholder: "Cari organisasi…", label: "Cari organisasi" }}
				filters={[
					{
						id: "status",
						label: "Status",
						options: [
							{ value: "active", label: "Aktif" },
							{ value: "suspended", label: "Ditangguhkan" },
						],
						value: status,
						onChange: (value) => {
							if (value === "" || value === "active" || value === "suspended") onFilter(search.settle(), value);
						},
					},
				]}
				onClear={() => onFilter(search.settle(""), "")}
				labels={filterLabels}
			/>
			{error && (
				<ContentState kind="error" title={error} description={data ? "Data terakhir tetap ditampilkan." : undefined} onRetry={onRetry} />
			)}
			{loading && !data && !error ? (
				<ContentState kind="loading" title="Memuat organisasi…" />
			) : data ? (
				<>
					{loading && (
						<p role="status" className="text-sm text-muted-foreground">
							Memperbarui organisasi… Data terakhir tetap ditampilkan.
						</p>
					)}
					{data.items.length ? (
						<Collection
							label="Daftar organisasi"
							items={data.items}
							getKey={(row) => row.id}
							getLink={organizationLink}
							columns={[
								{
									id: "name",
									header: "Organisasi",
									cell: (row) => (
										<span className="flex min-w-0 flex-col gap-0.5">
											<span>{row.name}</span>
											<span className="text-xs font-normal break-all text-muted-foreground">{row.slug}</span>
										</span>
									),
								},
								{
									id: "status",
									header: "Status",
									cell: (row) => <Badge variant="secondary">{row.status === "active" ? "Aktif" : "Ditangguhkan"}</Badge>,
								},
								{ id: "activeMembers", header: "Anggota aktif", align: "end", inlineLabel: true, cell: (row) => row.activeMemberCount },
								{ id: "members", header: "Semua keanggotaan", align: "end", inlineLabel: true, cell: (row) => row.memberCount },
								{ id: "owners", header: "Pemilik aktif", align: "end", inlineLabel: true, cell: (row) => row.activeOwnerCount },
								{ id: "created", header: "Dibuat", inlineLabel: true, cell: (row) => <DateTime value={row.createdAt} /> },
								{ id: "updated", header: "Diperbarui", inlineLabel: true, cell: (row) => <DateTime value={row.updatedAt} /> },
							]}
						/>
					) : (
						<ContentState
							kind="empty"
							title={filtered ? "Tidak ada organisasi yang cocok" : "Belum ada organisasi"}
							description={
								filtered
									? "Ubah atau kosongkan filter untuk mencari kembali."
									: "Buat organisasi beserta pemilik awalnya dalam satu langkah."
							}
							action={
								<Button variant="outline" onClick={filtered ? () => onFilter(search.settle(""), "") : onCreate}>
									{filtered ? "Kosongkan filter" : "Tambah organisasi"}
								</Button>
							}
						/>
					)}
					<div className="flex flex-wrap items-center justify-between gap-3 text-sm">
						<p>
							Halaman {data.pagination.page} · {data.pagination.total} organisasi
						</p>
						<div className="flex gap-2">
							<Button variant="outline" disabled={loading || data.pagination.page <= 1} onClick={() => onPage(data.pagination.page - 1)}>
								Sebelumnya
							</Button>
							<Button
								variant="outline"
								disabled={loading || BigInt(data.pagination.page) * BigInt(data.pagination.pageSize) >= BigInt(data.pagination.total)}
								onClick={() => onPage(data.pagination.page + 1)}
							>
								Berikutnya
							</Button>
						</div>
					</div>
					<p className="max-w-[65ch] text-xs text-muted-foreground">
						Anggota aktif menghitung keanggotaan aktif, termasuk pengguna yang dinonaktifkan. Pemilik aktif juga memerlukan pengguna aktif.
					</p>
				</>
			) : null}
		</PageContainer>
	);
}
