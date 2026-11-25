import { useEffect } from "react";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { usePrivateQuery } from "@bo/lib/api/session";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { useShellContext } from "../../../-components/shell-context";

export const Route = createFileRoute("/_authenticated/organizations/$organizationId")({
	component: OrganizationLayout,
});
function OrganizationLayout() {
	const { organizationId } = Route.useParams();
	const { setContext } = useShellContext();
	// Same query key as the child pages, so the profile is fetched once and cached.
	const organization = usePrivateQuery(
		["customer-organization", organizationId],
		async (t, signal) =>
			(await sdk.boV1GetOrganization({ client: t.client, signal, throwOnError: true, path: { id: organizationId } })).data,
	);
	const name = organization.data?.name;
	useEffect(() => {
		if (!name) return;
		setContext({ kind: "organization", id: organizationId, title: name });
		return () => setContext(null);
	}, [organizationId, name, setContext]);
	return <Outlet />;
}
