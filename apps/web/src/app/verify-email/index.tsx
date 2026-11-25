import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { VerifyEmailConfirm } from "./-components/verify-email-confirm";
import { RouteError, RoutePending } from "../-components/route-states";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { useLinkPage, useRedirectIfSignedIn } from "../../lib/session/hooks";

export const Route = createFileRoute("/verify-email/")({
	head: () =>
		createMetadata({
			title: "Verifikasi email",
			description: "Konfirmasi alamat email akun Zenith Anda.",
			pathname: "/verify-email",
			index: false,
		}),
	component: () => <ApiGate>{({ transport }) => <VerifyEmail t={transport} />}</ApiGate>,
});

function VerifyEmail({ t }: { t: AuthTransport }) {
	const signedIn = useRedirectIfSignedIn(t);
	const page = useLinkPage(t, "verify_email");
	if (signedIn) return <RoutePending />;
	if (page.failed) return <RouteError />;
	return <VerifyEmailConfirm state={page.state} confirm={page.form} onConfirm={() => page.confirm()} />;
}
