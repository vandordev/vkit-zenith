import { Link } from "@tanstack/react-router";
import { KeyRound, ShieldCheck, UsersRound } from "lucide-react";
import { Badge } from "@repo/ui/badge";
import { Button, buttonVariants } from "@repo/ui/button";
import { ErrorState, ErrorStateActions, ErrorStateContent, ErrorStateHeader, ErrorStateTitle } from "@repo/ui/error-state";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { StatsCard } from "@repo/ui/stats-card";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";
import { linkProps, orgLink } from "./navigation";

const roleLabels: Record<V1MyOrganization["role"], string> = { owner: "Owner", admin: "Admin", member: "Anggota" };

export type OrgOverviewProps = {
	organization: V1MyOrganization;
	/** Active member count; undefined while loading or after a failure. */
	memberCount?: number;
	membersState: "ready" | "pending" | "error";
	membersError?: string;
	onRetryMembers: () => void;
};

/** Ringkasan: the organization, the viewer's role and where to go next. Features add their own sections. */
export function OrgOverview({ organization, memberCount, membersState, membersError, onRetryMembers }: OrgOverviewProps) {
	const canDevelop = organization.permissions.includes("developers.manage");
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader
				title={organization.name}
				description={<span className="break-all">{organization.slug}</span>}
				status={<Badge variant="secondary">{organization.status === "active" ? "Aktif" : "Ditangguhkan"}</Badge>}
			/>
			{membersState === "error" ? (
				<ErrorState role="alert">
					<ErrorStateContent>
						<ErrorStateHeader>
							<ErrorStateTitle>{membersError ?? "Ringkasan tidak dapat dimuat"}</ErrorStateTitle>
						</ErrorStateHeader>
						<ErrorStateActions>
							<Button type="button" variant="outline" onClick={onRetryMembers}>
								Muat ulang
							</Button>
						</ErrorStateActions>
					</ErrorStateContent>
				</ErrorState>
			) : (
				<StatsCard
					ariaLabel="Ringkasan organisasi"
					items={[
						{ key: "role", title: "Peran Anda", value: { display: roleLabels[organization.role] }, icon: ShieldCheck },
						{
							key: "members",
							title: "Anggota aktif",
							value: { display: membersState === "pending" || memberCount === undefined ? "…" : String(memberCount) },
							icon: UsersRound,
						},
					]}
				/>
			)}
			<nav aria-label="Tindakan cepat" className="flex flex-wrap gap-2">
				<Link {...linkProps(orgLink(organization.slug, "/settings/members"))} className={buttonVariants({ variant: "outline" })}>
					<UsersRound aria-hidden="true" />
					Lihat anggota
				</Link>
				{canDevelop && (
					<Link {...linkProps(orgLink(organization.slug, "/settings/api-keys"))} className={buttonVariants({ variant: "outline" })}>
						<KeyRound aria-hidden="true" />
						Kelola API key
					</Link>
				)}
			</nav>
		</PageContainer>
	);
}
