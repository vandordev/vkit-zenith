import type { ReactNode } from "react";
import { Users, UserCheck, ShieldCheck } from "lucide-react";
import { Button } from "@repo/ui/button";
import { Badge } from "@repo/ui/badge";
import { DateTime } from "@repo/ui/date-time";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { StatsCard } from "@repo/ui/stats-card";
import { ContentState } from "../../../../-components/content-state";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
export function UserWorkspace({
	record,
	loading,
	error,
	emailActions,
	backLink,
	onRetry,
	onEdit,
	onEmail,
	onVerify,
	onStatus,
}: {
	record?: V1UserGetBody;
	loading?: boolean;
	error?: string;
	/** Email-action metadata section; its own data, states and paging belong to the adapter. */
	emailActions?: ReactNode;
	backLink: ReactNode;
	onRetry: () => void;
	onEdit: () => void;
	onEmail: () => void;
	onVerify: () => void;
	onStatus: () => void;
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
				!error && <ContentState kind="loading" title="Memuat pelanggan…" />
			) : (
				<>
					<PageHeader
						title={record.displayName}
						description={<span className="break-all">{record.email}</span>}
						status={<Badge variant="secondary">{record.status === "active" ? "Aktif" : "Dinonaktifkan"}</Badge>}
						actions={
							<Button disabled={loading} onClick={onEdit}>
								Ubah profil
							</Button>
						}
					/>
					<StatsCard
						ariaLabel="Keanggotaan pelanggan"
						items={[
							{ key: "total", title: "Semua keanggotaan", value: { display: record.membershipCount }, icon: Users },
							{ key: "active", title: "Keanggotaan aktif", value: { display: record.activeMembershipCount }, icon: UserCheck },
							{ key: "owner", title: "Kepemilikan aktif", value: { display: record.ownerMembershipCount }, icon: ShieldCheck },
						]}
					/>
					<div className="space-y-2 text-xs text-muted-foreground">
						<p>
							Dibuat <DateTime value={record.createdAt} />
						</p>
						<p>
							Diperbarui <DateTime value={record.updatedAt} />
						</p>
					</div>
					<section className="space-y-4 border-y border-border py-5" aria-label="Email pelanggan">
						<h2 className="text-lg font-semibold">Email pelanggan</h2>
						<p className="break-all text-sm">{record.email}</p>
						<p className="text-sm">
							{record.emailVerifiedAt ? (
								<>
									Terverifikasi pada <DateTime value={record.emailVerifiedAt} />
								</>
							) : (
								"Belum terverifikasi"
							)}
						</p>
						<div className="flex flex-wrap gap-3">
							<Button variant="outline" disabled={loading || !record.canChangeEmail} onClick={onEmail}>
								Ganti email
							</Button>
							<Button variant="outline" disabled={loading || !record.canVerifyEmail} onClick={onVerify}>
								Verifikasi email manual
							</Button>
						</div>
						{!record.canChangeEmail && <p className="text-xs text-muted-foreground">{record.emailBlockedReason}</p>}
						{!record.canVerifyEmail && <p className="text-xs text-muted-foreground">{record.verificationBlockedReason}</p>}
						<p className="max-w-[65ch] text-xs text-muted-foreground">Verifikasi bukan aktivasi pengguna atau pembuatan akses masuk.</p>
					</section>
					<section className="space-y-3" aria-label="Status pengguna">
						<h2 className="text-lg font-semibold">Status pengguna global</h2>
						<Button variant="outline" disabled={loading || !record.canChangeStatus} onClick={onStatus}>
							{record.status === "active" ? "Nonaktifkan pelanggan" : "Aktifkan pelanggan"}
						</Button>
						{!record.canChangeStatus && (
							<p className="text-sm text-muted-foreground">
								{record.statusBlockedReason} Kelola pemilik pengganti melalui organisasi terkait pada tab Organisasi.
							</p>
						)}
						<p className="max-w-[65ch] text-xs text-muted-foreground">
							Penonaktifan mempertahankan keanggotaan. Pengaktifan tidak memulihkan keanggotaan yang dihapus atau sesi yang telah dicabut.
						</p>
					</section>
					{emailActions}
				</>
			)}
		</PageContainer>
	);
}
