import { createFileRoute } from "@tanstack/react-router";
import { createMetadata } from "../../lib/metadata";
import { RegisterForm, type RegisterValues } from "./-components/register-form";
import { ApiGate } from "../../lib/api/provider";
import type { AuthTransport } from "../../lib/api/client";
import { register } from "../../lib/session/flows";
import { useAction, useRedirectIfSignedIn } from "../../lib/session/hooks";
import { RoutePending } from "../-components/route-states";
import { paths } from "../../lib/paths";

export const Route = createFileRoute("/register/")({
	head: () =>
		createMetadata({
			title: "Daftar",
			description: "Buat akun Zenith dan mulai menerima pembayaran lewat provider Anda sendiri.",
			pathname: "/register",
			index: true,
		}),
	component: () => <ApiGate>{({ transport }) => <Register t={transport} />}</ApiGate>,
});

function Register({ t }: { t: AuthTransport }) {
	const signedIn = useRedirectIfSignedIn(t);
	const [status, submit] = useAction(async (values: RegisterValues) => {
		const result = await register(t, values);
		if (result.kind !== "accepted") return result;
		window.location.assign(paths.checkEmail);
		return { kind: "pending" };
	});
	if (signedIn) return <RoutePending />;
	return <RegisterForm status={status} onSubmit={submit} />;
}
