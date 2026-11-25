import { Button } from "@repo/ui/button";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { StatusNotice } from "../../-components/status-notice";
import { paths } from "../../../lib/paths";

export type CheckEmailState = { kind: "restricted"; email: string; displayName: string } | { kind: "generic" };

export function CheckEmailPanel({
	state,
	resend,
	logout,
	onResend,
	onLogout,
}: {
	state: CheckEmailState;
	resend: FormStatus | { kind: "sent" };
	logout: FormStatus;
	onResend: () => void;
	onLogout: () => void;
}) {
	if (state.kind === "generic") {
		return (
			<AuthLayout
				title="Periksa email Anda"
				description="Jika pendaftaran dapat dilanjutkan, kami mengirim tautan ke email Anda. Jika Anda sudah punya akun, masuk atau atur ulang kata sandi."
			>
				<div className="flex flex-col gap-3">
					<a className="text-center text-sm font-medium underline underline-offset-4" href={paths.login}>
						Masuk
					</a>
					<a className="text-center text-sm font-medium underline underline-offset-4" href={paths.forgotPassword}>
						Lupa kata sandi
					</a>
				</div>
			</AuthLayout>
		);
	}
	const resendPending = resend.kind === "pending";
	const logoutPending = logout.kind === "pending";
	return (
		<AuthLayout
			title="Periksa email Anda"
			description={`Halo ${state.displayName}, kami mengirim tautan verifikasi ke ${state.email}. Buka tautan itu untuk melanjutkan.`}
		>
			<div className="flex flex-col gap-4">
				{resend.kind === "sent" ? (
					<p role="status" className="rounded-md border border-border bg-muted p-3 text-sm leading-relaxed">
						Email verifikasi dikirim ulang. Periksa kotak masuk Anda.
					</p>
				) : (
					<StatusNotice status={resend} />
				)}
				<Button isLoading={resendPending} onClick={() => !resendPending && onResend()}>
					Kirim ulang email
				</Button>
				<StatusNotice status={logout} />
				<Button variant="outline" isLoading={logoutPending} onClick={() => !logoutPending && onLogout()}>
					Keluar
				</Button>
			</div>
		</AuthLayout>
	);
}
