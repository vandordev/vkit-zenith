import { useEffect } from "react";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { usePrivateQuery } from "@bo/lib/api/session";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { useShellContext } from "../../../-components/shell-context";

export const Route = createFileRoute("/_authenticated/users/$userId")({ component: UserLayout });
function UserLayout() {
	const { userId } = Route.useParams();
	const { setContext } = useShellContext();
	// Same query key as the child pages, so the profile is fetched once and cached.
	const user = usePrivateQuery(
		["customer-user", userId],
		async (t, signal) => (await sdk.boV1GetUser({ client: t.client, signal, throwOnError: true, path: { id: userId } })).data,
	);
	const name = user.data?.displayName;
	// The cleanup clears the context before the next user (or leaving the section) publishes its own.
	useEffect(() => {
		if (!name) return;
		setContext({ kind: "user", id: userId, title: name });
		return () => setContext(null);
	}, [userId, name, setContext]);
	return <Outlet />;
}
