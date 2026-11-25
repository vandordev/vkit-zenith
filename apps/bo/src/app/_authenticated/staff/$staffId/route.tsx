import { useEffect } from "react";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { AccessState } from "../../../-components/access-state";
import { useShellContext } from "../../../-components/shell-context";
import * as sdk from "../../../../lib/api/generated/sdk.gen";
import { usePrivateQuery, useProtected } from "../../../../lib/api/session";

export const Route = createFileRoute("/_authenticated/staff/$staffId")({ component: StaffLayout });
function StaffLayout() {
	const { staff, epoch } = useProtected();
	// Staff context is superadmin-only: an admin never mounts a staff query or publishes context.
	return staff.role === "superadmin" ? <AuthorizedStaff key={`${staff.id}:${epoch}`} /> : <AccessState kind="denied" />;
}
function AuthorizedStaff() {
	const { staffId } = Route.useParams();
	const { setContext } = useShellContext();
	// Same query key as the child pages, so the account is fetched once and cached.
	const detail = usePrivateQuery(
		["staff-detail", staffId],
		async (t, signal) => (await sdk.boV1GetStaff({ client: t.client, signal, throwOnError: true, path: { id: staffId } })).data,
	);
	const name = detail.data?.displayName;
	// The cleanup clears the context before the next account (or leaving the section) publishes its own.
	useEffect(() => {
		if (!name) return;
		setContext({ kind: "staff", id: staffId, title: name });
		return () => setContext(null);
	}, [staffId, name, setContext]);
	return <Outlet />;
}
