import type { ReactNode } from "react";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { OrgShell } from "./org-shell";
import { linkProps } from "./navigation";
import { navigationFor, organization } from "./-fixtures";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";

const noop = () => {};

/** Story-only: a page inside the real OrgShell and resolver, so the active menu and mobile drawer can be inspected with it. */
export function ShellPreview({ children, org = organization() }: { children: ReactNode; org?: V1MyOrganization }) {
	const location = useLocation();
	const navigate = useNavigate();
	return (
		<OrgShell
			organization={org}
			organizations={[org]}
			navigation={navigationFor(org.slug, location.pathname)}
			userName="Rani Wulandari"
			onLogout={noop}
			theme="system"
			onThemeChange={noop}
			onNavigate={(link) => void navigate(linkProps(link))}
		>
			{children}
		</OrgShell>
	);
}
