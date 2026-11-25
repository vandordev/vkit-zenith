import type { ReactNode } from "react";
import { Building2, Users, ShieldCheck } from "lucide-react";
import { Button } from "@repo/ui/button";
import { Badge } from "@repo/ui/badge";
import { DateTime } from "@repo/ui/date-time";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { StatsCard } from "@repo/ui/stats-card";
import { ContentState } from "../../../../-components/content-state";
import type { V1OrganizationGetBody } from "@bo/lib/api/generated/types.gen";

export function OrganizationWorkspace({
	record,
	loading,
	error,
	backLink,
	onRetry,
	onEdit,
	onStatus,
	onOwners,
	children,
}: {
	record?: V1OrganizationGetBody;
	loading?: boolean;
	error?: string;
	backLink: ReactNode;
	onRetry: () => void;
	onEdit: () => void;
	onStatus: () => void;
	onOwners: () => void;
	children?: ReactNode;
}) {
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			{backLink}
			{error && (
				<ContentState
					kind="error"
					title={error}
					description={record ? "Profil terakhir tetap ditampilkan." : undefined}
					onRetry={onRetry}
				/>
			)}
			{!record ? (
				!error && <ContentState kind="loading" title="Memuat organisasi…" />
			) : (
				<>
					<PageHeader
						title={record.name}
						description={<span className="break-all">{record.slug}</span>}
						status={<Badge variant="secondary">{record.status === "active" ? "Aktif" : "Ditangguhkan"}</Badge>}
						actions={
							<Button disabled={loading} onClick={onEdit}>
								Ubah profil
							</Button>
						}
					/>
					{loading && (
						<p role="status" className="text-sm text-muted-foreground">
							Memperbarui profil organisasi…
						</p>
					)}
					<StatsCard
						ariaLabel="Ringkasan organisasi"
						items={[
							{ key: "members", title: "Semua keanggotaan", value: { display: record.memberCount }, icon: Building2 },
							{ key: "active", title: "Anggota aktif", value: { display: record.activeMemberCount }, icon: Users },
							{ key: "owners", title: "Pemilik aktif", value: { display: record.activeOwnerCount }, icon: ShieldCheck },
						]}
					/>
					<div className="flex flex-wrap items-center justify-between gap-4">
						<div className="space-y-1 text-xs text-muted-foreground">
							<p>
								Dibuat <DateTime value={record.createdAt} />
							</p>
							<p>
								Diperbarui <DateTime value={record.updatedAt} />
							</p>
						</div>
						<Button variant="outline" onClick={onOwners}>
							Lihat pemilik organisasi
						</Button>
					</div>
					<section className="space-y-3 border-y border-border py-5" aria-label="Status organisasi">
						<h2 className="font-semibold">Status organisasi</h2>
						<p className="max-w-[65ch] text-sm text-muted-foreground">
							Penangguhan berlaku untuk organisasi ini, bukan menonaktifkan pengguna global atau menghapus keanggotaan.
						</p>
						<Button variant="outline" disabled={loading || !record.canChangeStatus} onClick={onStatus}>
							{record.status === "active" ? "Tangguhkan organisasi" : "Aktifkan organisasi"}
						</Button>
						{!record.canChangeStatus && <p className="text-xs text-muted-foreground">{record.statusBlockedReason}</p>}
					</section>
					{children}
				</>
			)}
		</PageContainer>
	);
}
