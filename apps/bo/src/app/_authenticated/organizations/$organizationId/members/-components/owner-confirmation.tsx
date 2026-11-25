import { Select } from "@repo/ui/select";
import { useState } from "react";
import { Notice, type Feedback } from "../../../../../-components/common";
import { CustomerConfirmation } from "../../../-components/customer-confirmation";
import type { V1OrganizationMembershipGetBody, V1OrganizationMembershipRoleInputBody } from "@bo/lib/api/generated/types.gen";

type OwnerConfirmationProps = {
	record: V1OrganizationMembershipGetBody;
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onSubmit: (body: V1OrganizationMembershipRoleInputBody) => void;
	onReload: () => Promise<V1OrganizationMembershipGetBody | undefined>;
};
export function OwnerConfirmation(props: OwnerConfirmationProps) {
	return <OwnerForm key={`${props.record.organizationId}:${props.record.id}`} {...props} />;
}
function OwnerForm({ record, pending, feedback, onDirty, onSubmit, onReload }: OwnerConfirmationProps) {
	const [snapshot, setSnapshot] = useState(record);
	const [role, setRole] = useState<"owner" | "admin" | "member">(record.role === "owner" ? "member" : "owner");
	const [loading, setLoading] = useState(false);
	const [localFeedback, setLocalFeedback] = useState<Feedback>();
	const reload = async () => {
		setLoading(true);
		try {
			const fresh = await onReload();
			if (!fresh || fresh.id !== snapshot.id || fresh.organizationId !== snapshot.organizationId) {
				setLocalFeedback({ kind: "error", message: "Keanggotaan tidak tersedia dalam organisasi ini. Isian aman dipertahankan." });
				return;
			}
			setSnapshot(fresh);
			setLocalFeedback(undefined);
		} catch {
			setLocalFeedback({ kind: "error", message: "Keanggotaan tidak dapat dimuat ulang." });
		} finally {
			setLoading(false);
		}
	};
	return (
		<div className="space-y-4">
			<Notice feedback={localFeedback} />
			<Select
				label="Peran baru"
				value={role}
				data={[
					{ value: "owner", label: "Pemilik" },
					{ value: "admin", label: "Admin" },
					{ value: "member", label: "Anggota" },
				]}
				disabled={pending || loading}
				onValueChange={(value) => {
					if (value === "owner" || value === "admin" || value === "member") {
						setRole(value);
						onDirty(true);
					}
				}}
			/>
			<CustomerConfirmation
				targetKey={`${snapshot.id}:${snapshot.revision}`}
				allowed={snapshot.canChangeRole && role !== snapshot.role}
				blockedReason={snapshot.roleBlockedReason}
				pending={pending || loading}
				feedback={feedback}
				onDirty={onDirty}
				onReload={() => void reload()}
				onSubmit={(fields) => onSubmit({ ...fields, revision: snapshot.revision, role })}
				description={
					<>
						<p className="break-words">
							Ubah peran: <strong>{snapshot.userName}</strong> di <strong>{snapshot.organizationName}</strong>.
						</p>
						<p>
							{snapshot.role === "owner"
								? "Keanggotaan biasa tetap aktif. Penghapusan anggota adalah tindakan terpisah. Organisasi harus tetap mempunyai pemilik dengan pengguna aktif, termasuk ketika ditangguhkan."
								: "Pelanggan harus aktif dan keanggotaannya aktif."}
						</p>
					</>
				}
			/>
		</div>
	);
}
