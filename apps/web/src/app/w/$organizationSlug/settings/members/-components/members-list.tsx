import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@repo/ui/empty";
import { ErrorState, ErrorStateActions, ErrorStateContent, ErrorStateHeader, ErrorStateTitle } from "@repo/ui/error-state";
import { Loading } from "@repo/ui/loading";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { Collection } from "@repo/ui/collection";
import { Badge } from "@repo/ui/badge";
import { BoringAvatar } from "@repo/ui/boring-avatar";
import type { V1OrganizationMember } from "../../../../../../lib/api/generated/types.gen";

const roleLabels: Record<V1OrganizationMember["role"], string> = { owner: "Owner", admin: "Admin", member: "Anggota" };

export type MembersListProps = {
	state: "ready" | "pending" | "error";
	members?: readonly V1OrganizationMember[];
	errorMessage?: string;
	onRetry: () => void;
};

/** Active members only, read-only in v1: changes go through the Zenith team. */
export function MembersList({ state, members, errorMessage, onRetry }: MembersListProps) {
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader title="Anggota" description="Perubahan anggota melalui tim Zenith." />
			{state === "pending" && (
				<div className="flex min-h-32 items-center justify-center py-8" aria-busy="true">
					<Loading variant="skeleton" size={48} aria-label="Memuat anggota" />
				</div>
			)}
			{state === "error" && (
				<ErrorState role="alert">
					<ErrorStateContent>
						<ErrorStateHeader>
							<ErrorStateTitle>{errorMessage ?? "Anggota tidak dapat dimuat"}</ErrorStateTitle>
						</ErrorStateHeader>
						<ErrorStateActions>
							<Button type="button" variant="outline" onClick={onRetry}>
								Muat ulang
							</Button>
						</ErrorStateActions>
					</ErrorStateContent>
				</ErrorState>
			)}
			{state === "ready" && members?.length === 0 && (
				<Empty>
					<EmptyHeader>
						<EmptyTitle>Belum ada anggota aktif</EmptyTitle>
						<EmptyDescription>Hubungi tim Zenith untuk menambahkan anggota.</EmptyDescription>
					</EmptyHeader>
				</Empty>
			)}
			{state === "ready" && members && members.length > 0 && (
				<Collection
					label="Anggota organisasi"
					items={members}
					getKey={(member) => member.email}
					columns={[
						{ id: "name", header: "Nama", cell: (member) => <span className="font-medium">{member.name}</span> },
						{
							id: "avatar",
							header: "Avatar",
							slot: "media",
							cell: (member) => <BoringAvatar name={member.email} variant="beam" size={32} />,
						},
						{ id: "email", header: "Email", cell: (member) => <span className="break-all">{member.email}</span> },
						{
							id: "role",
							header: "Peran",
							cell: (member) => <Badge variant={member.role === "owner" ? "info-light" : "outline"}>{roleLabels[member.role]}</Badge>,
						},
						{ id: "joined", header: "Bergabung", inlineLabel: true, cell: (member) => <DateTime value={member.joinedAt} /> },
					]}
				/>
			)}
		</PageContainer>
	);
}
