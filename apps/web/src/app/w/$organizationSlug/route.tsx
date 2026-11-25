import { useEffect, useMemo } from "react";
import { createFileRoute, Outlet, useNavigate, useRouterState } from "@tanstack/react-router";
import { RouteError, RoutePending } from "../../-components/route-states";
import { useMemberships, useSignedIn } from "../-home";
import { OrgContextValue } from "./-components/org-context";
import { OrgShell } from "./-components/org-shell";
import { linkProps, resolveOrgNavigation, type OrgContext } from "./-components/navigation";
import { OrganizationNotFound } from "./-components/not-found";
import { appConfig } from "../../../lib/config";
import { createMetadata } from "../../../lib/metadata";
import { ApiGate, useIdentityEpoch } from "../../../lib/api/provider";
import type { AuthTransport } from "../../../lib/api/client";
import { logout } from "../../../lib/session/flows";
import { useAction } from "../../../lib/session/hooks";
import { writeLastOrganization } from "../../../lib/last-organization";
import { paths } from "../../../lib/paths";
import { useTheme } from "../../../lib/theme";

export const Route = createFileRoute("/w/$organizationSlug")({
	head: () => createMetadata({ title: "Dasbor", description: appConfig.defaultDescription, index: false }),
	component: () => <ApiGate>{({ transport }) => <Organization t={transport} />}</ApiGate>,
});

function Organization({ t }: { t: AuthTransport }) {
	const { organizationSlug } = Route.useParams();
	const epoch = useIdentityEpoch();
	const [theme, setTheme] = useTheme();
	const pathname = useRouterState({ select: (state) => state.location.pathname });
	const navigate = useNavigate();
	const { signedIn, redirecting, unavailable } = useSignedIn(t);
	const memberships = useMemberships(t, Boolean(signedIn));
	const organization = memberships.data?.find((item) => item.slug === organizationSlug);
	const [logoutStatus, submitLogout] = useAction(async () => {
		const result = await logout(t);
		if (result.kind !== "done") return result;
		window.location.assign(paths.login);
		return { kind: "pending" };
	});
	useEffect(() => {
		if (organization) writeLastOrganization(organization.slug);
	}, [organization]);
	const context = useMemo<OrgContext | undefined>(
		() => organization && { organization, can: (permission) => organization.permissions.includes(permission) },
		[organization],
	);
	if (unavailable || memberships.isError) return memberships.unauthorized ? <RoutePending /> : <RouteError />;
	if (!signedIn || redirecting || !memberships.data) return <RoutePending />;
	// Unknown slug, removed membership and non-member are one state.
	if (!organization || !context) return <OrganizationNotFound />;
	return (
		<OrgContextValue.Provider value={context}>
			{/* Keyed by the identity epoch: an identity change remounts every page and drops component state. */}
			<OrgShell
				key={epoch}
				organization={organization}
				organizations={memberships.data}
				navigation={resolveOrgNavigation({ pathname, slug: organization.slug })}
				userName={signedIn.displayName}
				logoutPending={logoutStatus.kind === "pending"}
				onLogout={submitLogout}
				theme={theme}
				onThemeChange={setTheme}
				onNavigate={(link) => void navigate(linkProps(link))}
			>
				<Outlet />
			</OrgShell>
		</OrgContextValue.Provider>
	);
}
