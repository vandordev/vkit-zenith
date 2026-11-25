import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { AuthTransport } from "../api/client";
import type { FormStatus } from "../../app/-components/form-status";
import type { LinkPageState } from "../../app/-components/link-outcome";
import { confirm, inspectForRoute, resolveHome, type Purpose } from "./flows";
import { paths } from "../paths";
import { singleFlight } from "./single-flight";
import { captureLinkToken } from "./token";

/** Single-flight submit. `run` returns the settled status; return `pending` when navigating away. */
export function useAction<A extends unknown[], R extends { kind: string } = never>(run: (...a: A) => Promise<FormStatus | R>) {
	const [status, setStatus] = useState<FormStatus | R>({ kind: "idle" });
	const latest = useRef(run);
	latest.current = run;
	const [submit] = useState(() =>
		singleFlight(async (...a: A) => {
			setStatus({ kind: "pending" });
			setStatus(await latest.current(...a));
		}),
	);
	return [status, submit] as const;
}

/** Captures the token before any Core call, inspects it, then confirms on explicit action. */
export function useLinkPage(t: AuthTransport, expected: Purpose) {
	const [token, setToken] = useState<string | null>();
	useEffect(() => setToken(captureLinkToken(window)), []);
	const [done, setDone] = useState<LinkPageState>();
	const inspect = useQuery({
		// the token never enters the key; it is held in memory only
		queryKey: ["inspect", expected, t.epoch()],
		enabled: Boolean(token) && !done,
		gcTime: 0,
		queryFn: () => inspectForRoute(t, token as string, expected),
	});
	const [form, submit] = useAction(async (token: string, values?: { password?: string; telephoneDigits?: string }) => {
		const result = await confirm(t, token, values);
		if (result.kind !== "outcome") return result;
		setDone(result);
		return { kind: "idle" };
	});
	const state: LinkPageState =
		done ?? (token === null ? { kind: "outcome", outcome: "unavailable" } : (inspect.data ?? { kind: "loading" }));
	return {
		state,
		form,
		failed: inspect.isError,
		confirm: (values?: { password?: string; telephoneDigits?: string }) => token && submit(token, values),
	};
}

/** Auth pages (login, register, ...) send a visitor who already has a full session straight to their workspace. */
export function useRedirectIfSignedIn(t: AuthTransport) {
	const home = useQuery({ queryKey: ["home", t.epoch()], queryFn: () => resolveHome(t), gcTime: 0 });
	const signedIn = home.data?.kind === "signed-in";
	useEffect(() => {
		if (signedIn) window.location.replace(paths.home);
	}, [signedIn]);
	return signedIn;
}
