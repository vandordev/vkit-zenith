import { useState } from "react";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { FieldError, StatusNotice, fieldError } from "../../-components/status-notice";
import { paths } from "../../../lib/paths";

export function ForgotPasswordForm({ status, onSubmit }: { status: FormStatus; onSubmit: (values: { email: string }) => void }) {
	const [email, setEmail] = useState("");
	const pending = status.kind === "pending";
	const emailError = fieldError(status, "email");
	return (
		<AuthLayout
			title="Lupa kata sandi"
			description="Masukkan email Anda. Kami akan mengirim tautan untuk mengatur kata sandi."
			footer={<a href={paths.login}>Kembali ke halaman masuk</a>}
		>
			<form
				className="flex flex-col gap-5"
				onSubmit={(event) => {
					event.preventDefault();
					if (!pending) onSubmit({ email });
				}}
			>
				<StatusNotice status={status} />
				<div className="flex flex-col gap-2">
					<Input
						label="Email"
						name="email"
						type="email"
						autoComplete="email"
						required
						maxLength={254}
						value={email}
						onChange={(event) => setEmail(event.target.value)}
						disabled={pending}
						aria-invalid={emailError ? true : undefined}
						aria-describedby={emailError ? "forgot-email-error" : undefined}
					/>
					<FieldError id="forgot-email-error" message={emailError} />
				</div>
				<Button type="submit" isLoading={pending}>
					Kirim tautan
				</Button>
			</form>
		</AuthLayout>
	);
}

export function RequestAccepted() {
	return (
		<AuthLayout
			title="Periksa email Anda"
			description="Jika alamat terdaftar, kami mengirim tautan. Periksa email Anda."
			footer={<a href={paths.login}>Kembali ke halaman masuk</a>}
		>
			{null}
		</AuthLayout>
	);
}
