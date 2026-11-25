import { useState } from "react";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { InputPassword } from "@repo/ui/input-password";
import { InputPhone } from "@repo/ui/input-phone";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { FieldError, StatusNotice, fieldError, useSecretField } from "../../-components/status-notice";
import { paths } from "../../../lib/paths";

export type RegisterValues = { displayName: string; email: string; telephoneDigits: string; password: string };

export function RegisterForm({
	status,
	defaults,
	onSubmit,
}: {
	status: FormStatus;
	defaults?: { displayName?: string; email?: string; telephone?: string };
	onSubmit: (values: RegisterValues) => void;
}) {
	const [displayName, setDisplayName] = useState(defaults?.displayName ?? "");
	const [email, setEmail] = useState(defaults?.email ?? "");
	const [telephoneDigits, setTelephoneDigits] = useState(defaults?.telephone ?? "");
	const [password, setPassword] = useSecretField(status);
	const pending = status.kind === "pending";
	const errors = {
		displayName: fieldError(status, "displayName"),
		email: fieldError(status, "email"),
		telephone: fieldError(status, "telephone"),
		password: fieldError(status, "password"),
	};
	return (
		<AuthLayout
			title="Buat akun Zenith"
			description="Isi data berikut. Kami akan mengirim tautan verifikasi ke email Anda."
			footer={<a href={paths.login}>Sudah punya akun? Masuk</a>}
		>
			<form
				className="flex flex-col gap-5"
				onSubmit={(event) => {
					event.preventDefault();
					if (!pending) onSubmit({ displayName, email, telephoneDigits, password });
				}}
			>
				<StatusNotice status={status} />
				<div className="flex flex-col gap-2">
					<Input
						label="Nama lengkap"
						name="displayName"
						autoComplete="name"
						required
						value={displayName}
						onChange={(event) => setDisplayName(event.target.value)}
						disabled={pending}
						aria-invalid={errors.displayName ? true : undefined}
						aria-describedby={errors.displayName ? "register-name-error" : undefined}
					/>
					<FieldError id="register-name-error" message={errors.displayName} />
				</div>
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
						aria-invalid={errors.email ? true : undefined}
						aria-describedby={errors.email ? "register-email-error" : undefined}
					/>
					<FieldError id="register-email-error" message={errors.email} />
				</div>
				<div className="flex flex-col gap-2">
					<InputPhone
						label="Nomor telepon"
						defaultCountry="ID"
						locale="id"
						countrySelectLabel="Kode negara"
						required
						value={telephoneDigits}
						onValueChange={(digits) => setTelephoneDigits(digits ?? "")}
						disabled={pending}
						aria-invalid={errors.telephone ? true : undefined}
						aria-describedby={errors.telephone ? "register-telephone-error" : undefined}
					/>
					<FieldError id="register-telephone-error" message={errors.telephone} />
				</div>
				<div className="flex flex-col gap-2">
					<InputPassword
						label="Kata sandi"
						name="password"
						autoComplete="new-password"
						required
						maxLength={512}
						value={password}
						onChange={(event) => setPassword(event.target.value)}
						disabled={pending}
						aria-invalid={errors.password ? true : undefined}
						aria-describedby={errors.password ? "register-password-error" : undefined}
						showPasswordLabel="Tampilkan kata sandi"
						hidePasswordLabel="Sembunyikan kata sandi"
					/>
					<FieldError id="register-password-error" message={errors.password} />
				</div>
				<Button type="submit" isLoading={pending}>
					Daftar
				</Button>
			</form>
		</AuthLayout>
	);
}
