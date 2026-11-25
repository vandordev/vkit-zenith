import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { LoginForm } from "./-components/login-form";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { login } from "../../lib/session/flows";
import { useAction, useRedirectIfSignedIn } from "../../lib/session/hooks";
import { RoutePending } from "../-components/route-states";
import { paths } from "../../lib/paths";

export const Route = createFileRoute("/login/")({
	head: () =>
		createMetadata({
			title: "Masuk",
			description: "Masuk ke dasbor Zenith untuk mengelola produk, pembayaran, dan langganan bisnis Anda.",
			pathname: "/login",
			index: true,
		}),
	component: () => <ApiGate>{({ transport }) => <Login t={transport} />}</ApiGate>,
});

function Login({ t }: { t: AuthTransport }) {
	const signedIn = useRedirectIfSignedIn(t);
	const [status, submit] = useAction(async (values: { email: string; password: string }) => {
		const result = await login(t, values);
		if (result.kind !== "full" && result.kind !== "restricted") return result;
		window.location.assign(result.kind === "full" ? paths.home : paths.checkEmail);
		return { kind: "pending" };
	});
	if (signedIn) return <RoutePending />;
	return <LoginForm status={status} onSubmit={submit} />;
}
