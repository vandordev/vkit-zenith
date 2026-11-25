import { useState } from "react";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
import type { Feedback } from "../../../../-components/common";
export type UserActionProps = {
	record: V1UserGetBody;
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onReload: () => Promise<V1UserGetBody | undefined>;
};
export function useUserAction(record: V1UserGetBody, onReload: UserActionProps["onReload"]) {
	const [snapshot, setSnapshot] = useState(record);
	const [loading, setLoading] = useState(false);
	const [localFeedback, setLocalFeedback] = useState<Feedback>();
	const reload = async () => {
		setLoading(true);
		try {
			const fresh = await onReload();
			if (!fresh || fresh.id !== snapshot.id) {
				setLocalFeedback({ kind: "error", message: "Pelanggan tidak dapat dimuat. Isian aman dipertahankan." });
				return;
			}
			setSnapshot(fresh);
			setLocalFeedback(undefined);
		} catch {
			setLocalFeedback({ kind: "error", message: "Pelanggan tidak dapat dimuat ulang." });
		} finally {
			setLoading(false);
		}
	};
	return { snapshot, loading, localFeedback, reload };
}
