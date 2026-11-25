import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { ResetPasswordForm } from "./-components/reset-password-form";
import { RouteError } from "../-components/route-states";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { useLinkPage } from "../../lib/session/hooks";

export const Route = createFileRoute("/reset-password/")({
	head: () =>
		createMetadata({
			title: "Atur ulang kata sandi",
			description: "Buat kata sandi baru untuk akun Zenith Anda.",
			pathname: "/reset-password",
			index: false,
		}),
	component: () => <ApiGate>{({ transport }) => <ResetPassword t={transport} />}</ApiGate>,
});

function ResetPassword({ t }: { t: AuthTransport }) {
	const page = useLinkPage(t, "reset_password");
	if (page.failed) return <RouteError />;
	return <ResetPasswordForm state={page.state} form={page.form} onSubmit={page.confirm} />;
}
