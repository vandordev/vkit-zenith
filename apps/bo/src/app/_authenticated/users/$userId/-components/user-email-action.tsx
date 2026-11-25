import { useState } from "react";
import { Input } from "@repo/ui/input";
import { Notice } from "../../../../-components/common";
import { CustomerConfirmation } from "../../../organizations/-components/customer-confirmation";
import { useUserAction, type UserActionProps } from "./use-user-action";
import type { V1UserEmailInputBody } from "@bo/lib/api/generated/types.gen";
type Props = UserActionProps & { onSubmit: (body: V1UserEmailInputBody) => void };
export function UserEmailAction(props: Props) {
	return <EmailForm key={props.record.id} {...props} />;
}
function EmailForm({ record, pending, feedback, onDirty, onReload, onSubmit }: Props) {
	const { snapshot, loading, localFeedback, reload } = useUserAction(record, onReload);
	const [email, setEmail] = useState("");
	const unchanged = email.trim().toLowerCase() === snapshot.email.trim().toLowerCase();
	return (
		<div className="space-y-4">
			<Notice feedback={localFeedback} />
			<CustomerConfirmation
				targetKey={`${snapshot.id}:${snapshot.revision}`}
				allowed={snapshot.canChangeEmail && !unchanged}
				blockedReason={unchanged ? "Email baru harus berbeda dari email saat ini." : snapshot.emailBlockedReason}
				pending={pending || loading}
				feedback={feedback}
				onDirty={onDirty}
				onReload={() => void reload()}
				onSubmit={(fields) => onSubmit({ ...fields, email, revision: snapshot.revision })}
				description={
					<>
						<p>
							Ganti email <span className="break-all font-medium">{snapshot.email}</span> milik <strong>{snapshot.displayName}</strong>.
						</p>
						<p>
							Email baru belum terverifikasi. Sesi, token email dan otorisasi pelanggan yang berlaku dicabut. Identitas dan keanggotaan
							tetap sama.
						</p>
					</>
				}
			>
				<Input
					label="Email baru"
					name="email"
					type="email"
					value={email}
					required
					maxLength={254}
					disabled={pending || loading || !snapshot.canChangeEmail}
					onChange={(e) => {
						setEmail(e.target.value);
						onDirty(true);
					}}
				/>
			</CustomerConfirmation>
		</div>
	);
}
