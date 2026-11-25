import type { ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Badge } from "@repo/ui/badge";
import { BoringAvatar } from "@repo/ui/boring-avatar";
import { Collection } from "@repo/ui/collection";
import { DateTime } from "@repo/ui/date-time";
import { FilterBar } from "@repo/ui/filter-bar";
import { ContentState } from "../../../../../-components/content-state";
import { filterLabels } from "../../../../../-components/filter-labels";
import { useSearchDraft } from "../../../../../-components/search-draft";
import type { V1OrganizationMembershipListBody, V1OrganizationMembershipListItem } from "@bo/lib/api/generated/types.gen";

export type MembershipFilters = { q: string; status: "" | "active" | "removed"; ownership: "" | "owner" | "member" };
export function MembershipCollection({
	data,
	filters,
	loading,
	error,
	onFilter,
	onPage,
	onRetry,
	onAdd,
	onSelect,
	userLink,
	navigationKey = "",
}: {
	data?: V1OrganizationMembershipListBody;
	filters: MembershipFilters;
	loading?: boolean;
	error?: string;
	onFilter: (value: MembershipFilters) => void;
	onPage: (page: number) => void;
	onRetry: () => void;
	onAdd: () => void;
	onSelect: (row: V1OrganizationMembershipListItem) => void;
	userLink: (row: V1OrganizationMembershipListItem) => ReactNode;
	navigationKey?: string;
}) {
	const search = useSearchDraft(filters.q, (q) => onFilter({ ...filters, q }), navigationKey);
	const filtered = Boolean(filters.q || filters.status || filters.ownership);
	return (
		<section className="min-w-0 space-y-5" aria-label="Keanggotaan organisasi">
			<div className="flex flex-wrap items-center justify-between gap-3">
				<h2 className="text-xl font-semibold">Anggota organisasi</h2>
				<Button onClick={onAdd}>Tambah anggota</Button>
			</div>
			{data && (
				<p className="text-sm text-muted-foreground">
					Seluruh keanggotaan: {data.summary.total} · Aktif: {data.summary.active} · Dihapus: {data.summary.removed} · Keanggotaan pemilik
					aktif: {data.summary.owners}
				</p>
			)}
			<FilterBar
				search={{ value: search.draft, onChange: search.change, placeholder: "Cari anggota…", label: "Cari anggota" }}
				filters={[
					{
						id: "status",
						label: "Keanggotaan",
						options: [
							{ value: "active", label: "Aktif" },
							{ value: "removed", label: "Dihapus" },
						],
						value: filters.status,
						onChange: (v) => {
							if (v === "" || v === "active" || v === "removed") onFilter({ ...filters, q: search.settle(), status: v });
						},
					},
					{
						id: "ownership",
						label: "Kepemilikan",
						options: [
							{ value: "owner", label: "Pemilik" },
							{ value: "member", label: "Bukan pemilik" },
						],
						value: filters.ownership,
						onChange: (v) => {
							if (v === "" || v === "owner" || v === "member") onFilter({ ...filters, q: search.settle(), ownership: v });
						},
					},
				]}
				onClear={() => onFilter({ q: search.settle(""), status: "", ownership: "" })}
				labels={filterLabels}
			/>
			{error && (
				<ContentState
					kind="error"
					title={error}
					description={data ? "Keanggotaan terakhir tetap ditampilkan." : undefined}
					onRetry={onRetry}
				/>
			)}
			{!data ? (
				!error && <ContentState kind="loading" title="Memuat anggota…" />
			) : (
				<>
					{loading && (
						<p role="status" className="text-sm text-muted-foreground">
							Memperbarui anggota…
						</p>
					)}
					{data.items.length ? (
						// The row's action is managing the membership, so the row is not one link; the name links to the user.
						<Collection
							label="Keanggotaan organisasi"
							items={data.items}
							getKey={(row) => row.id}
							columns={[
								{
									id: "user",
									header: "Pelanggan",
									cell: (row) => (
										<span className="flex min-w-0 flex-col gap-0.5">
											<span className="font-medium">{userLink(row)}</span>
											<span className="text-xs break-all text-muted-foreground">{row.userEmail}</span>
										</span>
									),
								},
								{
									id: "avatar",
									header: "Avatar",
									slot: "media",
									cell: (row) => <BoringAvatar name={row.userId} variant="beam" size={32} />,
								},
								{
									id: "userStatus",
									header: "Pengguna global",
									inlineLabel: true,
									cell: (row) => (row.userStatus === "active" ? "Aktif" : "Dinonaktifkan"),
								},
								{
									id: "status",
									header: "Keanggotaan",
									cell: (row) => <Badge variant="secondary">{row.status === "active" ? "Aktif" : "Dihapus"}</Badge>,
								},
								{
									id: "role",
									header: "Peran",
									inlineLabel: true,
									cell: (row) => (row.role === "owner" ? "Pemilik" : row.role === "admin" ? "Admin" : "Anggota"),
								},
								{ id: "updated", header: "Diperbarui", inlineLabel: true, cell: (row) => <DateTime value={row.updatedAt} /> },
								{
									id: "actions",
									header: "Tindakan",
									slot: "actions",
									cell: (row) => (
										<Button variant="outline" size="sm" disabled={loading} onClick={() => onSelect(row)}>
											Kelola anggota
										</Button>
									),
								},
							]}
						/>
					) : (
						<ContentState
							kind="empty"
							title={filtered ? "Tidak ada anggota yang cocok" : "Belum ada keanggotaan"}
							description={
								filtered
									? "Ubah atau kosongkan filter. Daftar ini dapat memiliki halaman lain."
									: "Tambahkan pelanggan aktif sebagai anggota biasa."
							}
						/>
					)}
					<div className="flex flex-wrap items-center justify-between gap-3 text-sm">
						<p>
							Halaman {data.pagination.page} · {data.pagination.total} keanggotaan
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
						Keanggotaan aktif tetap dihitung saat pengguna dinonaktifkan. Daftar pemilik menggunakan filter dan paginasi, bukan daftar
						lengkap dalam pratinjau.
					</p>
				</>
			)}
		</section>
	);
}
