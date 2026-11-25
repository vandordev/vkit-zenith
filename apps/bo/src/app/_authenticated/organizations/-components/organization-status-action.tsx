import { useState } from "react";
import { Notice, type Feedback } from "../../../-components/common";
import { CustomerConfirmation } from "./customer-confirmation";
import type { V1OrganizationGetBody, V1OrganizationStatusInputBody } from "@bo/lib/api/generated/types.gen";
type Props = {
	record: V1OrganizationGetBody;
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onSubmit: (body: V1OrganizationStatusInputBody) => void;
	onReload: () => Promise<V1OrganizationGetBody | undefined>;
};
export function OrganizationStatusAction(props: Props) {
	return <StatusForm key={props.record.id} {...props} />;
}
function StatusForm({ record, pending, feedback, onDirty, onSubmit, onReload }: Props) {
	const [snapshot, setSnapshot] = useState(record);
	const [loading, setLoading] = useState(false);
	const [localFeedback, setLocalFeedback] = useState<Feedback>();
	const reload = async () => {
		setLoading(true);
		try {
			const fresh = await onReload();
			if (!fresh || fresh.id !== snapshot.id) {
				setLocalFeedback({ kind: "error", message: "Organisasi tidak dapat dimuat. Isian aman dipertahankan." });
				return;
			}
			setSnapshot(fresh);
			setLocalFeedback(undefined);
		} catch {
			setLocalFeedback({ kind: "error", message: "Organisasi tidak dapat dimuat ulang." });
		} finally {
			setLoading(false);
		}
	};
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
					onSubmit({ ...fields, revision: snapshot.revision, status: snapshot.status === "active" ? "suspended" : "active" })
				}
				description={
					<>
						<p>
							{snapshot.status === "active" ? "Tangguhkan" : "Aktifkan"} organisasi <strong>{snapshot.name}</strong>.
						</p>
						<p>
							Tindakan ini tidak menonaktifkan pengguna global, menghapus anggota atau mengubah kepemilikan.
						</p>
					</>
				}
			/>
		</div>
	);
}
