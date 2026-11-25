import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditForm, type Feedback } from "../../../../-components/common";
import type { V1UserGetBody, V1UserUpdateInputBody } from "@bo/lib/api/generated/types.gen";
type Props = {
	record: V1UserGetBody;
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onSubmit: (body: V1UserUpdateInputBody) => void;
	onReload: () => Promise<V1UserGetBody | undefined>;
};
export function UserProfile(props: Props) {
	return <ProfileForm key={props.record.id} {...props} />;
}
function ProfileForm({ record, pending, feedback, onDirty, onSubmit, onReload }: Props) {
	const [snapshot, setSnapshot] = useState(record);
	const [version, setVersion] = useState(0);
	const [dirty, setDirty] = useState(false);
	const [loading, setLoading] = useState(false);
	const [localFeedback, setLocalFeedback] = useState<Feedback>();
	const reload = async () => {
		if (pending || loading || (dirty && !window.confirm("Buang perubahan yang belum disimpan dan muat ulang pelanggan?"))) return;
		setLoading(true);
		try {
			const fresh = await onReload();
			if (!fresh || fresh.id !== snapshot.id) {
				setLocalFeedback({ kind: "error", message: "Pelanggan tidak dapat dimuat. Isian tetap dipertahankan." });
				return;
			}
			setSnapshot(fresh);
			setVersion((v) => v + 1);
			setDirty(false);
			onDirty(false);
			setLocalFeedback(undefined);
		} catch {
			setLocalFeedback({ kind: "error", message: "Pelanggan tidak dapat dimuat. Isian tetap dipertahankan." });
		} finally {
			setLoading(false);
		}
	};
	return (
		<div className="space-y-4">
			<EditForm
				key={version}
				fields={[{ name: "displayName", label: "Nama pelanggan", maxLength: 200 }]}
				initial={{ displayName: snapshot.displayName }}
				pending={pending || loading}
				feedback={localFeedback ?? feedback}
				onDirty={(v) => {
					setDirty(v);
					onDirty(v);
				}}
				onSubmit={(values) => onSubmit({ displayName: values.displayName ?? "", revision: snapshot.revision })}
			/>
			{feedback?.kind === "error" && (
				<Button variant="outline" disabled={pending || loading} onClick={() => void reload()}>
					Muat ulang revisi pelanggan
				</Button>
			)}
			<p className="text-xs text-muted-foreground">Profil hanya mengubah nama. Email dan status memakai tindakan terpisah.</p>
		</div>
	);
}
