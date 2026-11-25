import { useState } from "react";
import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { InputPassword } from "@repo/ui/input-password";
import { InputPhone } from "@repo/ui/input-phone";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { LinkStateGate, type LinkPageState } from "../../-components/link-outcome";
import { FieldError, StatusNotice, fieldError, useSecretField } from "../../-components/status-notice";

export function SetPasswordForm({
	state,
	form,
	onSubmit,
}: {
	state: LinkPageState;
	form: FormStatus;
	onSubmit: (values: { telephoneDigits: string; password: string }) => void;
}) {
	const [telephoneDigits, setTelephoneDigits] = useState("");
	const [password, setPassword] = useSecretField(form);
	if (state.kind !== "ready") return <LinkStateGate state={state} />;
	const pending = form.kind === "pending";
	const errors = { telephone: fieldError(form, "telephone"), password: fieldError(form, "password") };
	return (
		<AuthLayout title="Atur kata sandi" description="Lengkapi nomor telepon dan buat kata sandi pertama Anda.">
			<form
				className="flex flex-col gap-5"
				onSubmit={(event) => {
					event.preventDefault();
					if (!pending) onSubmit({ telephoneDigits, password });
				}}
			>
				<StatusNotice status={form} />
				{state.displayName ? (
					<p className="break-words text-sm">
						Akun untuk <span className="font-medium">{state.displayName}</span>
					</p>
				) : null}
				<p className="text-sm text-muted-foreground">
					Tautan berlaku sampai <DateTime value={state.expiresAt} locale="id" />.
				</p>
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
						aria-describedby={errors.telephone ? "set-password-telephone-error" : undefined}
					/>
					<FieldError id="set-password-telephone-error" message={errors.telephone} />
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
						aria-describedby={errors.password ? "set-password-error" : undefined}
						showPasswordLabel="Tampilkan kata sandi"
						hidePasswordLabel="Sembunyikan kata sandi"
					/>
					<FieldError id="set-password-error" message={errors.password} />
				</div>
				<Button type="submit" isLoading={pending}>
					Simpan kata sandi
				</Button>
			</form>
		</AuthLayout>
	);
}
