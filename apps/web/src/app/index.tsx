import { createFileRoute, Navigate } from "@tanstack/react-router";
import { RouteError, RoutePending } from "./-components/route-states";
import { useSignedIn } from "./w/-home";
import { appConfig } from "../lib/config";
import { createMetadata } from "../lib/metadata";
import { ApiGate } from "../lib/api/provider";
import type { AuthTransport } from "../lib/api/client";

export const Route = createFileRoute("/")({
	head: () => createMetadata({ title: appConfig.defaultTitle, description: appConfig.defaultDescription, pathname: "/", index: false }),
	component: () => <ApiGate>{({ transport }) => <Root t={transport} />}</ApiGate>,
});

/** `/` has no page: anonymous → login and restricted → check-email (via useSignedIn), full session → /w. */
function Root({ t }: { t: AuthTransport }) {
	const { signedIn, unavailable } = useSignedIn(t);
	if (unavailable) return <RouteError />;
	return signedIn ? <Navigate to="/w" replace /> : <RoutePending />;
}
