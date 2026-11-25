import type { ReactElement } from "react";
import { DateTime } from "@repo/ui/date-time";
import { Collection } from "@repo/ui/collection";
import type { V1UserMembershipListBody, V1UserMembershipListItem } from "@bo/lib/api/generated/types.gen";
import { listState, OrgSection, type SectionState } from "../../../../../-components/org-section";
import { Pager } from "../../../../../-components/pager";

const roles = { owner: "Pemilik", admin: "Admin", member: "Anggota" } as const;

/** Memberships of one user (extracted from the user overview); each organization links to its member workspace. */
export function UserOrganizations({
	memberships,
	state,
	error,
	loading,
	organizationLink,
	onPage,
	onRetry,
}: {
	memberships?: V1UserMembershipListBody;
	state: SectionState;
	error?: string;
	loading?: boolean;
	/** The organization's member page, a router `Link` without children; the whole row opens it. */
	organizationLink: (row: V1UserMembershipListItem) => ReactElement;
	onPage: (page: number) => void;
	onRetry: () => void;
}) {
	const items = memberships?.items ?? [];
	const pagination = memberships?.pagination;
	return (
		<OrgSection
			title="Keanggotaan organisasi"
			description="Status pengguna, organisasi dan keanggotaan berbeda. Kelola anggota melalui organisasi terkait."
			state={listState(state, items.length, pagination?.total)}
			pager={pagination && <Pager pagination={pagination} noun="keanggotaan" loading={loading} onPage={onPage} />}
			error={error}
			onRetry={onRetry}
			emptyTitle="Belum ada keanggotaan"
			emptyDescription="Pelanggan dapat berdiri sendiri tanpa organisasi."
		>
			<Collection
				label="Organisasi pelanggan"
				items={items}
				getKey={(row) => row.id}
				getLink={organizationLink}
				columns={[
					{
						id: "organization",
						header: "Organisasi",
						cell: (row) => (
							<span className="flex min-w-0 flex-col gap-0.5">
								<span>{row.organizationName}</span>
								<span className="text-xs font-normal break-all text-muted-foreground">{row.organizationSlug}</span>
							</span>
						),
					},
					{
						id: "organizationStatus",
						header: "Status organisasi",
						inlineLabel: true,
						cell: (row) => (row.organizationStatus === "active" ? "Aktif" : "Ditangguhkan"),
					},
					{ id: "membership", header: "Keanggotaan", inlineLabel: true, cell: (row) => (row.status === "active" ? "Aktif" : "Dihapus") },
					{ id: "role", header: "Kepemilikan", inlineLabel: true, cell: (row) => roles[row.role as keyof typeof roles] ?? row.role },
					{ id: "updated", header: "Diperbarui", inlineLabel: true, cell: (row) => <DateTime value={row.updatedAt} /> },
				]}
			/>
		</OrgSection>
	);
}
