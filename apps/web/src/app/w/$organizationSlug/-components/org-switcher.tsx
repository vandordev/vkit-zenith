import { Link } from "@tanstack/react-router";
import { Check, ChevronsUpDown, Settings } from "lucide-react";
import type { SVGProps } from "react";
import { Badge } from "@repo/ui/badge";
import { BoringAvatar } from "@repo/ui/boring-avatar";
import { Button } from "@repo/ui/button";
import { Dropdown, type DropdownEntry } from "@repo/ui/dropdown";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";
import { linkProps } from "./navigation";

const roleLabels: Record<V1MyOrganization["role"], string> = { owner: "Owner", admin: "Admin", member: "Anggota" };

// Keeps non-current rows aligned with the checkmark.
const Blank = (props: SVGProps<SVGSVGElement>) => <svg {...props} />;

function Identity({ organization, collapsed }: { organization: V1MyOrganization; collapsed: boolean }) {
	return (
		<>
			{/* Real logo rendering is M4: a generated marble from the stable id is the placeholder for now. */}
			<BoringAvatar name={organization.id} variant="marble" size={28} square className="overflow-hidden rounded-md" />
			<span
				aria-hidden={collapsed || undefined}
				className={`min-w-0 flex-1 text-left text-xs transition-opacity duration-(--motion-fast) motion-reduce:transition-none${collapsed ? " opacity-0" : ""}`}
			>
				<span className="block truncate text-sm font-medium">{organization.name}</span>
			</span>
		</>
	);
}

/** Always a switcher, even with one organization (its row still offers settings). There is no "create organization" in v1. */
export function OrgSwitcher({
	organization,
	organizations,
	collapsed = false,
}: {
	organization: V1MyOrganization;
	organizations: readonly V1MyOrganization[];
	collapsed?: boolean;
}) {
	// The 28px initial sits on the rail's 24px icon axis in both states; the header wrapper already draws the divider.
	const rowClass =
		"flex h-auto min-h-12 w-full items-center justify-start gap-2 overflow-hidden ps-0.5 pe-2 py-2 has-[>svg]:ps-0.5 has-[>svg]:pe-2";
	const items: DropdownEntry[] = organizations.map((item) => ({
		id: item.id,
		label: item.name,
		icon: item.id === organization.id ? Check : Blank,
		link: (
			<Link
				{...linkProps({
					to: "/w/$organizationSlug",
					params: { organizationSlug: item.slug },
				})}
				aria-current={item.id === organization.id ? "true" : undefined}
			/>
		),
		badge: <Badge variant="outline">{roleLabels[item.role]}</Badge>,
		endAction: {
			label: `Pengaturan ${item.name}`,
			icon: Settings,
			link: (
				<Link
					{...linkProps({
						to: "/w/$organizationSlug/settings/profile",
						params: { organizationSlug: item.slug },
					})}
				/>
			),
		},
	}));
	return (
		<Dropdown
			contentProps={{ side: collapsed ? "right" : "bottom", align: "start", className: "w-60" }}
			trigger={
				<Button variant="ghost" aria-label={`Organisasi: ${organization.name}. Ganti organisasi`} className={rowClass}>
					<Identity organization={organization} collapsed={collapsed} />
					<ChevronsUpDown aria-hidden="true" className={`size-3.5 shrink-0${collapsed ? " hidden" : ""}`} />
				</Button>
			}
			items={[{ id: "organizations", type: "group", label: "Organisasi Anda", items }]}
		/>
	);
}
