import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { SetPasswordForm } from "./-components/set-password-form";
import { RouteError } from "../-components/route-states";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { useLinkPage } from "../../lib/session/hooks";

export const Route = createFileRoute("/set-password/")({
	head: () =>
		createMetadata({
			title: "Buat kata sandi",
			description: "Buat kata sandi untuk akun Zenith Anda.",
			pathname: "/set-password",
			index: false,
		}),
	component: () => <ApiGate>{({ transport }) => <SetPassword t={transport} />}</ApiGate>,
});

function SetPassword({ t }: { t: AuthTransport }) {
	const page = useLinkPage(t, "setup_password");
	if (page.failed) return <RouteError />;
	return <SetPasswordForm state={page.state} form={page.form} onSubmit={page.confirm} />;
}
