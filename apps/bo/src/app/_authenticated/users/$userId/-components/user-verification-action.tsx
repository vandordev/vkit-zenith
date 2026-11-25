import { Notice } from "../../../../-components/common";
import { CustomerConfirmation } from "../../../organizations/-components/customer-confirmation";
import { useUserAction, type UserActionProps } from "./use-user-action";
import type { V1UserVerifyInputBody } from "@bo/lib/api/generated/types.gen";
type Props = UserActionProps & { onSubmit: (body: V1UserVerifyInputBody) => void };
export function UserVerificationAction(props: Props) {
	return <VerificationForm key={props.record.id} {...props} />;
}
function VerificationForm({ record, pending, feedback, onDirty, onReload, onSubmit }: Props) {
	const { snapshot, loading, localFeedback, reload } = useUserAction(record, onReload);
	return (
		<div className="space-y-4">
			<Notice feedback={localFeedback} />
			<CustomerConfirmation
				targetKey={`${snapshot.id}:${snapshot.revision}`}
				allowed={snapshot.canVerifyEmail}
				blockedReason={snapshot.verificationBlockedReason}
				pending={pending || loading}
				feedback={feedback}
				onDirty={onDirty}
				onReload={() => void reload()}
				onSubmit={(fields) => onSubmit({ ...fields, revision: snapshot.revision })}
				description={
					<>
						<p>
							Verifikasi manual email saat ini: <strong className="break-all">{snapshot.email}</strong>.
						</p>
						<p>
							Gunakan hanya setelah pemeriksaan operasional. Ini tindakan terpisah, bukan pengiriman email verifikasi. Perubahan email
							bersamaan membatalkan revisi ini. Verifikasi tidak mengaktifkan pengguna yang dinonaktifkan atau membuat akses masuk.
						</p>
					</>
				}
			/>
		</div>
	);
}
