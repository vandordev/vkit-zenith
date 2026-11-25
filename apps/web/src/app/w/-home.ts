import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import type { AuthTransport } from "../../lib/api/client";
import { errorStatus, listMyOrganizations, orgQueryKey } from "../../lib/api/organization";
import { resolveHome } from "../../lib/session/flows";
import { paths } from "../../lib/paths";

/**
 * Session gate shared by /w and every /w/... direct URL: reads the session and renews a full
 * one once per page load (gcTime 0), and sends anonymous or restricted visitors to their page.
 */
export function useSignedIn(t: AuthTransport) {
	const home = useQuery({ queryKey: ["home", t.epoch()], queryFn: () => resolveHome(t), gcTime: 0 });
	const target = home.data?.kind === "anonymous" ? paths.login : home.data?.kind === "restricted" ? paths.checkEmail : undefined;
	useEffect(() => {
		if (target) window.location.replace(target);
	}, [target]);
	return {
		signedIn: home.data?.kind === "signed-in" ? home.data : undefined,
		redirecting: Boolean(target),
		unavailable: home.data?.kind === "unavailable",
	};
}

/** Active memberships, scoped to the identity epoch; a 401 means the session ended and goes to login. */
export function useMemberships(t: AuthTransport, enabled: boolean) {
	const query = useQuery({
		queryKey: orgQueryKey(t, "me", "none", "organizations"),
		queryFn: () => listMyOrganizations(t),
		enabled,
		gcTime: 0,
	});
	const unauthorized = errorStatus(query.error) === 401;
	useEffect(() => {
		if (unauthorized) window.location.replace(paths.login);
	}, [unauthorized]);
	return { ...query, unauthorized };
}
