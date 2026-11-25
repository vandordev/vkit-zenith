import type { ReactElement } from "react";
import { Users, UserCheck, UserX, MailCheck, MailQuestion } from "lucide-react";
import { Button } from "@repo/ui/button";
import { BoringAvatar } from "@repo/ui/boring-avatar";
import { Collection } from "@repo/ui/collection";
import { DateTime } from "@repo/ui/date-time";
import { FilterBar } from "@repo/ui/filter-bar";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { StatsCard } from "@repo/ui/stats-card";
import { ContentState } from "../../../-components/content-state";
import { filterLabels } from "../../../-components/filter-labels";
import { useSearchDraft } from "../../../-components/search-draft";
import type { V1UserListBody, V1UserListItem } from "@bo/lib/api/generated/types.gen";
export type UserFilters = { q: string; status: "" | "active" | "disabled"; verification: "" | "verified" | "unverified" };
export function UserCollection({
	data,
	filters,
	loading,
	error,
	onCreate,
	onFilter,
	onPage,
	onRetry,
	userLink,
	navigationKey = "",
}: {
	data?: V1UserListBody;
	filters: UserFilters;
	loading?: boolean;
	error?: string;
	onCreate: () => void;
	onFilter: (filters: UserFilters) => void;
	onPage: (page: number) => void;
	onRetry: () => void;
	/** The user's page, a router `Link` without children; the whole row opens it. */
	userLink: (row: V1UserListItem) => ReactElement;
	navigationKey?: string;
}) {
	const search = useSearchDraft(filters.q, (q) => onFilter({ ...filters, q }), navigationKey);
	const filtered = Boolean(filters.q || filters.status || filters.verification);
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader
				title="Pengguna pelanggan"
				description="Identitas pelanggan global, terpisah dari akun staf."
				actions={<Button onClick={onCreate}>Tambah pelanggan</Button>}
			/>
			<StatsCard
				ariaLabel="Ringkasan seluruh pelanggan"
				items={[
					{ key: "total", title: "Total pelanggan", value: { display: data?.summary.total ?? "—" }, icon: Users },
					{ key: "active", title: "Aktif", value: { display: data?.summary.active ?? "—" }, icon: UserCheck },
					{ key: "disabled", title: "Dinonaktifkan", value: { display: data?.summary.disabled ?? "—" }, icon: UserX },
					{ key: "verified", title: "Email terverifikasi", value: { display: data?.summary.verified ?? "—" }, icon: MailCheck },
					{
						key: "unverified",
						title: "Email belum terverifikasi",
						value: { display: data?.summary.unverified ?? "—" },
						icon: MailQuestion,
					},
				]}
			/>
			<FilterBar
				search={{ value: search.draft, onChange: search.change, placeholder: "Cari nama atau email…", label: "Cari nama atau email" }}
				filters={[
					{
						id: "status",
						label: "Status pengguna",
						options: [
							{ value: "active", label: "Aktif" },
							{ value: "disabled", label: "Dinonaktifkan" },
						],
						value: filters.status,
						onChange: (v) => {
							if (v === "" || v === "active" || v === "disabled") onFilter({ ...filters, q: search.settle(), status: v });
						},
					},
					{
						id: "verification",
						label: "Verifikasi email",
						options: [
							{ value: "verified", label: "Terverifikasi" },
							{ value: "unverified", label: "Belum terverifikasi" },
						],
						value: filters.verification,
						onChange: (v) => {
							if (v === "" || v === "verified" || v === "unverified") onFilter({ ...filters, q: search.settle(), verification: v });
						},
					},
				]}
				onClear={() => onFilter({ q: search.settle(""), status: "", verification: "" })}
				labels={filterLabels}
			/>
			{error && (
				<ContentState kind="error" title={error} description={data ? "Data terakhir tetap ditampilkan." : undefined} onRetry={onRetry} />
			)}
			{!data ? (
				!error && <ContentState kind="loading" title="Memuat pelanggan…" />
			) : (
				<>
					{loading && (
						<p role="status" className="text-sm text-muted-foreground">
							Memperbarui pelanggan…
						</p>
					)}
					{data.items.length ? (
						<Collection
							label="Daftar pelanggan"
							items={data.items}
							getKey={(row) => row.id}
							getLink={userLink}
							columns={[
								{
									id: "name",
									header: "Pelanggan",
									cell: (row) => (
										<span className="flex min-w-0 flex-col gap-0.5">
											<span>{row.displayName}</span>
											<span className="text-xs font-normal break-all text-muted-foreground">{row.email}</span>
										</span>
									),
								},
								{ id: "avatar", header: "Avatar", slot: "media", cell: (row) => <BoringAvatar name={row.id} variant="beam" size={32} /> },
								{
									id: "status",
									header: "Status global",
									inlineLabel: true,
									cell: (row) => (row.status === "active" ? "Aktif" : "Dinonaktifkan"),
								},
								{
									id: "email",
									header: "Email",
									inlineLabel: true,
									cell: (row) =>
										row.emailVerifiedAt ? (
											<span className="inline-flex flex-wrap items-baseline gap-x-1.5">
												Terverifikasi
												<DateTime value={row.emailVerifiedAt} className="text-xs whitespace-nowrap text-muted-foreground" />
											</span>
										) : (
											"Belum terverifikasi"
										),
								},
								{
									id: "activeMemberships",
									header: "Keanggotaan aktif",
									align: "end",
									inlineLabel: true,
									cell: (row) => row.activeMembershipCount,
								},
								{ id: "memberships", header: "Semua keanggotaan", align: "end", inlineLabel: true, cell: (row) => row.membershipCount },
								{ id: "owners", header: "Kepemilikan aktif", align: "end", inlineLabel: true, cell: (row) => row.ownerMembershipCount },
								{ id: "created", header: "Dibuat", inlineLabel: true, cell: (row) => <DateTime value={row.createdAt} /> },
								{ id: "updated", header: "Diperbarui", inlineLabel: true, cell: (row) => <DateTime value={row.updatedAt} /> },
							]}
						/>
					) : (
						<ContentState
							kind="empty"
							title={filtered ? "Tidak ada pelanggan yang cocok" : "Belum ada pelanggan"}
							description={filtered ? "Ubah atau kosongkan filter." : "Tambahkan profil pelanggan tanpa kata sandi atau akses masuk."}
							action={
								<Button
									variant="outline"
									onClick={filtered ? () => onFilter({ q: search.settle(""), status: "", verification: "" }) : onCreate}
								>
									{filtered ? "Kosongkan filter" : "Tambah pelanggan"}
								</Button>
							}
						/>
					)}
					<div className="flex flex-wrap items-center justify-between gap-3 text-sm">
						<p>
							Halaman {data.pagination.page} · {data.pagination.total} pelanggan
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
						Keanggotaan aktif dihitung terpisah dari status pengguna dan organisasi. Kepemilikan aktif menghitung keanggotaan aktif yang
						ditandai sebagai pemilik.
					</p>
				</>
			)}
		</PageContainer>
	);
}
