import { Notice } from "../../../../-components/common";
import { CustomerConfirmation } from "../../../organizations/-components/customer-confirmation";
import { useUserAction, type UserActionProps } from "./use-user-action";
import type { V1UserStatusInputBody } from "@bo/lib/api/generated/types.gen";
type Props = UserActionProps & { onSubmit: (body: V1UserStatusInputBody) => void };
export function UserStatusAction(props: Props) {
	return <StatusForm key={props.record.id} {...props} />;
}
function StatusForm({ record, pending, feedback, onDirty, onReload, onSubmit }: Props) {
	const { snapshot, loading, localFeedback, reload } = useUserAction(record, onReload);
	return (
		<div className="space-y-4">
			<Notice feedback={localFeedback} />
			<CustomerConfirmation
				targetKey={`${snapshot.id}:${snapshot.revision}`}
				allowed={snapshot.canChangeStatus}
				blockedReason={snapshot.statusBlockedReason}
				pending={pending || loading}
				feedback={feedback}
				onDirty={onDirty}
				onReload={() => void reload()}
				onSubmit={(fields) =>
					onSubmit({ ...fields, revision: snapshot.revision, status: snapshot.status === "active" ? "disabled" : "active" })
				}
				description={
					<>
						<p>
							{snapshot.status === "active" ? "Nonaktifkan" : "Aktifkan"} pelanggan <strong>{snapshot.displayName}</strong>.
						</p>
						<p>
							{snapshot.status === "active"
								? "Keanggotaan dipertahankan, tetapi sesi, token email dan otorisasi pelanggan yang berlaku dicabut secara atomik. Organisasi harus mempunyai pemilik aktif pengganti sebelum pemilik terakhir dinonaktifkan. Kelola pengganti melalui organisasi terkait, tanpa transfer otomatis."
								: "Keanggotaan aktif tetap aktif. Keanggotaan yang dihapus tidak dipulihkan dan sesi maupun token yang telah dicabut tidak hidup kembali. Pengaktifan tidak membuat kata sandi atau akses masuk."}
						</p>
					</>
				}
			/>
		</div>
	);
}
