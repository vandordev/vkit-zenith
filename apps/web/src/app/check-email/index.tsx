import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { useQuery } from "@tanstack/react-query";
import { CheckEmailPanel } from "./-components/check-email-panel";
import { RouteError, RoutePending } from "../-components/route-states";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { checkEmailState, logout, resend } from "../../lib/session/flows";
import { useAction, useRedirectIfSignedIn } from "../../lib/session/hooks";
import { paths } from "../../lib/paths";

export const Route = createFileRoute("/check-email/")({
	head: () =>
		createMetadata({
			title: "Periksa email",
			description: "Periksa email Anda untuk menyelesaikan verifikasi akun.",
			pathname: "/check-email",
			index: false,
		}),
	component: () => <ApiGate>{({ transport }) => <CheckEmail t={transport} />}</ApiGate>,
});

function CheckEmail({ t }: { t: AuthTransport }) {
	const signedIn = useRedirectIfSignedIn(t);
	const state = useQuery({ queryKey: ["check-email", t.epoch()], queryFn: () => checkEmailState(t), gcTime: 0 });
	const [resendStatus, onResend] = useAction<[], { kind: "sent" }>(async () => {
		const result = await resend(t);
		return result.kind === "done" ? { kind: "sent" } : result;
	});
	const [logoutStatus, onLogout] = useAction(async () => {
		const result = await logout(t);
		if (result.kind !== "done") return result;
		window.location.assign(paths.login);
		return { kind: "pending" };
	});
	if (signedIn || !state.data) return <RoutePending />;
	if (state.data.kind === "unavailable") return <RouteError />;
	return <CheckEmailPanel state={state.data} resend={resendStatus} logout={logoutStatus} onResend={onResend} onLogout={onLogout} />;
}
