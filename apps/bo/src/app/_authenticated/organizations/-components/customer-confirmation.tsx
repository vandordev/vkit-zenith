import { useEffect, useState, type ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { InputPassword } from "@repo/ui/input-password";
import { Notice, type Feedback } from "../../../-components/common";
import type { V1OrganizationMembershipRoleInputBody } from "@bo/lib/api/generated/types.gen";

// Feature composition, not a primitive or transport DTO. Only the per-action
// callbacks receive these fields; there is no reusable unlock/confirmation token.
export function CustomerConfirmation({
	targetKey,
	allowed,
	blockedReason,
	pending,
	feedback,
	description,
	children,
	onDirty,
	onSubmit,
	onReload,
}: {
	targetKey: string;
	allowed: boolean;
	blockedReason: string;
	pending?: boolean;
	feedback?: Feedback;
	description: ReactNode;
	children?: ReactNode;
	onDirty: (value: boolean) => void;
	onSubmit: (fields: Pick<V1OrganizationMembershipRoleInputBody, "reason" | "currentPassword">) => void;
	onReload?: () => void;
}) {
	const [reason, setReason] = useState("");
	const [password, setPassword] = useState("");
	useEffect(() => {
		setPassword("");
	}, [targetKey, allowed]);
	return (
		<form
			className="min-w-0 space-y-5"
			onInvalidCapture={() => setPassword("")}
			onSubmit={(e) => {
				e.preventDefault();
				if (pending || !allowed) return;
				const currentPassword = password;
				setPassword("");
				onSubmit({ reason, currentPassword });
			}}
		>
			<Notice feedback={feedback} />
			<div className="max-w-[65ch] space-y-3 text-sm leading-relaxed text-muted-foreground">{description}</div>
			{children}
			{allowed ? (
				<>
					<Input
						label="Alasan tindakan"
						name="reason"
						value={reason}
						required
						maxLength={1024}
						disabled={pending}
						onChange={(e) => {
							setReason(e.target.value);
							onDirty(true);
						}}
					/>
					<InputPassword
						label="Kata sandi Superadmin saat ini"
						name="currentPassword"
						value={password}
						required
						maxLength={512}
						autoComplete="current-password"
						disabled={pending}
						onChange={(e) => {
							setPassword(e.target.value);
							onDirty(true);
						}}
						showPasswordLabel="Tampilkan kata sandi"
						hidePasswordLabel="Sembunyikan kata sandi"
					/>
					<p className="text-xs text-muted-foreground">
						Alasan dicatat untuk audit. Jangan masukkan kata sandi atau rahasia token. Konfirmasi berlaku hanya untuk tindakan ini.
					</p>
					<Button type="submit" isLoading={pending}>
						Konfirmasi tindakan
					</Button>
				</>
			) : (
				<p role="status" className="text-sm">
					{blockedReason || "Tindakan tidak tersedia."}
				</p>
			)}
			{feedback?.kind === "error" && onReload && (
				<Button
					type="button"
					variant="outline"
					disabled={pending}
					onClick={() => {
						setPassword("");
						onReload();
					}}
				>
					Muat ulang detail dan revisi
				</Button>
			)}
		</form>
	);
}
