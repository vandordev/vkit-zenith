import type { ReactNode } from "react";
import { AuthLayout } from "./auth-layout";
import { paths } from "../../lib/paths";

export type LinkOutcomeKind = "unavailable" | "login-required" | "completed-login" | "completed-signed-in";

export type LinkPageState =
	| { kind: "loading" }
	| { kind: "ready"; expiresAt: string; displayName?: string }
	| { kind: "cross-browser" }
	| { kind: "outcome"; outcome: LinkOutcomeKind };

const outcomes: Record<LinkOutcomeKind, { title: string; description: string; footer: ReactNode }> = {
	// invalid, expired, used, wrong purpose and missing token all look the same by design
	unavailable: {
		title: "Tautan tidak tersedia",
		description: "Tautan ini tidak valid, sudah kedaluwarsa, atau sudah digunakan. Minta tautan baru untuk melanjutkan.",
		footer: <a href={paths.forgotPassword}>Minta tautan baru</a>,
	},
	"login-required": {
		title: "Masuk untuk melanjutkan",
		description: "Tautan dibuka di browser lain. Masuk di browser ini, atau minta tautan baru.",
		footer: (
			<>
				<a href={paths.login}>Masuk</a> · <a href={paths.forgotPassword}>Lupa kata sandi</a>
			</>
		),
	},
	"completed-login": {
		title: "Berhasil",
		description: "Perubahan Anda tersimpan. Silakan masuk dengan kata sandi Anda.",
		footer: <a href={paths.login}>Masuk</a>,
	},
	"completed-signed-in": {
		title: "Email terverifikasi",
		description: "Email Anda sudah terverifikasi dan Anda sudah masuk.",
		footer: <a href={paths.home}>Lanjut ke beranda</a>,
	},
};

export function LinkOutcome({ kind }: { kind: LinkOutcomeKind }) {
	const { title, description, footer } = outcomes[kind];
	return (
		<AuthLayout title={title} description={description} footer={footer}>
			{null}
		</AuthLayout>
	);
}

/** Renders every non-ready link state; returns null once the page itself should render. */
export function LinkStateGate({ state }: { state: LinkPageState }) {
	switch (state.kind) {
		case "loading":
			return (
				<AuthLayout title="Memeriksa tautan…">
					<p role="status" className="text-center text-sm text-muted-foreground">
						Mohon tunggu sebentar.
					</p>
				</AuthLayout>
			);
		case "cross-browser":
			return <LinkOutcome kind="login-required" />;
		case "outcome":
			return <LinkOutcome kind={state.outcome} />;
		case "ready":
			return null;
	}
}
