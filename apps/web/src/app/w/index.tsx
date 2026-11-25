import { createFileRoute, Navigate } from "@tanstack/react-router";
import { z } from "zod";
import { RouteError, RoutePending } from "../-components/route-states";
import { NoOrganization } from "./-components/no-organization";
import { useMemberships, useSignedIn } from "./-home";
import { appConfig } from "../../lib/config";
import { createMetadata } from "../../lib/metadata";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { logout } from "../../lib/session/flows";
import { useAction } from "../../lib/session/hooks";
import { readLastOrganization } from "../../lib/last-organization";
import { paths } from "../../lib/paths";

export const Route = createFileRoute("/w/")({
	head: () => createMetadata({ title: "Dasbor", description: appConfig.defaultDescription, index: false }),
	validateSearch: z.object({ mode: z.literal("test").optional().catch(undefined) }),
	component: () => <ApiGate>{({ transport }) => <Entry t={transport} />}</ApiGate>,
});

/** Entry table: 0 memberships → NoOrganization, 1 → its dashboard, several → last opened (else first by name). `mode` is preserved. */
function Entry({ t }: { t: AuthTransport }) {
	const { mode } = Route.useSearch();
	const { signedIn, redirecting, unavailable } = useSignedIn(t);
	const memberships = useMemberships(t, Boolean(signedIn));
	const [status, submit] = useAction(async () => {
		const result = await logout(t);
		if (result.kind !== "done") return result;
		window.location.assign(paths.login);
		return { kind: "pending" };
	});
	if (unavailable || memberships.isError) return memberships.unauthorized ? <RoutePending /> : <RouteError />;
	if (!signedIn || redirecting || !memberships.data) return <RoutePending />;
	if (memberships.data.length === 0) return <NoOrganization displayName={signedIn.displayName} logout={status} onLogout={submit} />;
	const last = readLastOrganization();
	const target = memberships.data.find((item) => item.slug === last) ?? memberships.data[0];
	if (!target) return <RouteError />;
	return <Navigate to="/w/$organizationSlug" params={{ organizationSlug: target.slug }} search={mode ? { mode } : {}} replace />;
}
