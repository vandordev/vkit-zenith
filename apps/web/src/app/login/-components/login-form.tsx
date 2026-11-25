import { useState } from "react";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { InputPassword } from "@repo/ui/input-password";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { FieldError, StatusNotice, fieldError, useSecretField } from "../../-components/status-notice";
import { paths } from "../../../lib/paths";

export type LoginValues = { email: string; password: string };

export function LoginFields({
	status,
	defaultEmail = "",
	onSubmit,
}: {
	status: FormStatus;
	defaultEmail?: string;
	onSubmit: (values: LoginValues) => void;
}) {
	const [email, setEmail] = useState(defaultEmail);
	const [password, setPassword] = useSecretField(status);
	const pending = status.kind === "pending";
	const emailError = fieldError(status, "email");
	const passwordError = fieldError(status, "password");
	return (
		<form
			className="flex flex-col gap-5"
			onSubmit={(event) => {
				event.preventDefault();
				if (!pending) onSubmit({ email, password });
			}}
		>
			<StatusNotice status={status} />
			<div className="flex flex-col gap-2">
				<Input
					label="Email"
					className="h-12"
					name="email"
					type="email"
					autoComplete="username"
					required
					maxLength={254}
					value={email}
					onChange={(event) => setEmail(event.target.value)}
					disabled={pending}
					aria-invalid={emailError ? true : undefined}
					aria-describedby={emailError ? "login-email-error" : undefined}
				/>
				<FieldError id="login-email-error" message={emailError} />
			</div>
			<div className="flex flex-col gap-2">
				<InputPassword
					label="Kata sandi"
					className="h-12"
					name="password"
					autoComplete="current-password"
					required
					maxLength={512}
					value={password}
					onChange={(event) => setPassword(event.target.value)}
					disabled={pending}
					aria-invalid={passwordError ? true : undefined}
					aria-describedby={passwordError ? "login-password-error" : undefined}
					showPasswordLabel="Tampilkan kata sandi"
					hidePasswordLabel="Sembunyikan kata sandi"
				/>
				<FieldError id="login-password-error" message={passwordError} />
			</div>
			<Button type="submit" size="lg" isLoading={pending}>
				Masuk
			</Button>
		</form>
	);
}

export const loginFooter = (
	<div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3">
		<a href={paths.forgotPassword}>Lupa kata sandi?</a>
		<a href={paths.register}>Belum punya akun? Daftar</a>
	</div>
);

export function LoginForm(props: { status: FormStatus; defaultEmail?: string; onSubmit: (values: LoginValues) => void }) {
	return (
		<AuthLayout title="Masuk ke akun Zenith" description="Gunakan email dan kata sandi Anda." footer={loginFooter}>
			<LoginFields {...props} />
		</AuthLayout>
	);
}
