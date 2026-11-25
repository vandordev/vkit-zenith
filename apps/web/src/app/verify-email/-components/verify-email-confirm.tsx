import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { LinkStateGate, type LinkPageState } from "../../-components/link-outcome";
import { StatusNotice } from "../../-components/status-notice";

export function VerifyEmailConfirm({ state, confirm, onConfirm }: { state: LinkPageState; confirm: FormStatus; onConfirm: () => void }) {
	if (state.kind !== "ready") return <LinkStateGate state={state} />;
	const pending = confirm.kind === "pending";
	return (
		<AuthLayout title="Verifikasi email" description="Konfirmasi untuk menyelesaikan verifikasi alamat email Anda.">
			<div className="flex flex-col gap-4">
				<StatusNotice status={confirm} />
				<p className="text-sm text-muted-foreground">
					Tautan berlaku sampai <DateTime value={state.expiresAt} locale="id" />.
				</p>
				<Button isLoading={pending} onClick={() => !pending && onConfirm()}>
					Verifikasi email
				</Button>
			</div>
		</AuthLayout>
	);
}
