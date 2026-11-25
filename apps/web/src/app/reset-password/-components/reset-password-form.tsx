import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { InputPassword } from "@repo/ui/input-password";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { LinkStateGate, type LinkPageState } from "../../-components/link-outcome";
import { FieldError, StatusNotice, fieldError, useSecretField } from "../../-components/status-notice";

export function ResetPasswordForm({
	state,
	form,
	onSubmit,
}: {
	state: LinkPageState;
	form: FormStatus;
	onSubmit: (values: { password: string }) => void;
}) {
	const [password, setPassword] = useSecretField(form);
	if (state.kind !== "ready") return <LinkStateGate state={state} />;
	const pending = form.kind === "pending";
	const passwordError = fieldError(form, "password");
	return (
		<AuthLayout title="Atur ulang kata sandi" description="Buat kata sandi baru untuk akun Zenith Anda.">
			<form
				className="flex flex-col gap-5"
				onSubmit={(event) => {
					event.preventDefault();
					if (!pending) onSubmit({ password });
				}}
			>
				<StatusNotice status={form} />
				<p className="text-sm text-muted-foreground">
					Tautan berlaku sampai <DateTime value={state.expiresAt} locale="id" />.
				</p>
				<div className="flex flex-col gap-2">
					<InputPassword
						label="Kata sandi baru"
						name="password"
						autoComplete="new-password"
						required
						maxLength={512}
						value={password}
						onChange={(event) => setPassword(event.target.value)}
						disabled={pending}
						aria-invalid={passwordError ? true : undefined}
						aria-describedby={passwordError ? "reset-password-error" : undefined}
						showPasswordLabel="Tampilkan kata sandi"
						hidePasswordLabel="Sembunyikan kata sandi"
					/>
					<FieldError id="reset-password-error" message={passwordError} />
				</div>
				<Button type="submit" isLoading={pending}>
					Simpan kata sandi
				</Button>
			</form>
		</AuthLayout>
	);
}
