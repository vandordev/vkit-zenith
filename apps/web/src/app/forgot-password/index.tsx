import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { ForgotPasswordForm, RequestAccepted } from "./-components/forgot-password-form";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { requestPasswordLink } from "../../lib/session/flows";
import { RoutePending } from "../-components/route-states";
import { useAction, useRedirectIfSignedIn } from "../../lib/session/hooks";

export const Route = createFileRoute("/forgot-password/")({
	head: () =>
		createMetadata({
			title: "Lupa kata sandi",
			description: "Minta tautan untuk mengatur ulang kata sandi akun Zenith Anda.",
			pathname: "/forgot-password",
			index: false,
		}),
	component: () => <ApiGate>{({ transport }) => <ForgotPassword t={transport} />}</ApiGate>,
});

function ForgotPassword({ t }: { t: AuthTransport }) {
	const signedIn = useRedirectIfSignedIn(t);
	const [status, submit] = useAction<[{ email: string }], { kind: "done" }>(({ email }) => requestPasswordLink(t, email));
	if (signedIn) return <RoutePending />;
	return status.kind === "done" ? <RequestAccepted /> : <ForgotPasswordForm status={status} onSubmit={submit} />;
}
