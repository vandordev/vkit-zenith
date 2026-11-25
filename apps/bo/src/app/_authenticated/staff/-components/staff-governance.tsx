import { useState } from "react";
import { Button } from "@repo/ui/button";
import { Notice, type Feedback } from "../../../-components/common";
import { ContentState } from "../../../-components/content-state";
import { StaffEditor } from "./staff-editor";
import type { BoV1GetStaffResponse } from "../../../../lib/api/generated/types.gen";
export function StaffGovernance({
	staff,
	creating,
	pending,
	feedback,
	onSubmit,
	onDirty,
	onReload,
	loadError,
	refreshing,
	onRetry,
}: {
	staff?: BoV1GetStaffResponse;
	creating?: boolean;
	pending?: boolean;
	feedback?: Feedback;
	onSubmit: (values: Record<string, string>, revision?: string, replacing?: boolean) => void;
	onDirty: (dirty: boolean) => void;
	onReload?: () => Promise<BoV1GetStaffResponse | undefined>;
	loadError?: string;
	refreshing?: boolean;
	onRetry?: () => void;
}) {
	const [snapshot, setSnapshot] = useState(staff);
	const [replacing, setReplacing] = useState(false);
	const [dirty, setDirty] = useState(false);
	const [key, setKey] = useState(0);
	const [reloading, setReloading] = useState(false);
	const [localError, setLocalError] = useState<Feedback>();
	const busy = pending || reloading;
	return (
		<div className="flex flex-col gap-5">
			{loadError && <ContentState kind="error" title={loadError} onRetry={onRetry} />}
			{refreshing && !loadError && (
				<p role="status" className="text-sm text-muted-foreground">
					Memperbarui staf… Isian dan revisi tetap dipertahankan.
				</p>
			)}
			<Notice feedback={localError} />
			<StaffEditor
				key={`${key}-${replacing}`}
				initial={
					snapshot
						? { email: snapshot.email, displayName: snapshot.displayName, role: snapshot.role, status: snapshot.status }
						: { role: "admin", status: "active" }
				}
				creating={creating}
				replacing={replacing}
				pending={busy}
				feedback={feedback}
				onDirty={(value) => {
					setDirty(value);
					onDirty(value);
				}}
				onSubmit={(values) => onSubmit(values, snapshot?.revision, replacing)}
			/>
			{snapshot && (
				<>
					<Button
						variant="outline"
						className="h-auto w-full min-w-0 justify-start !whitespace-normal break-words text-left"
						disabled={busy}
						onClick={() => {
							if (dirty && !window.confirm("Ganti tugas dan buang perubahan staf yang belum disimpan?")) return;
							setReplacing(!replacing);
							setDirty(false);
							onDirty(false);
						}}
					>
						{replacing ? "Kembali ke profil staf" : `Ganti kata sandi ${snapshot.displayName}`}
					</Button>
					{onReload && (
						<Button
							variant="outline"
							disabled={busy}
							onClick={async () => {
								if (dirty && !window.confirm("Muat ulang dan buang perubahan staf yang belum disimpan?")) return;
								setReloading(true);
								try {
									const next = await onReload();
									if (next) {
										setSnapshot(next);
										setKey((k) => k + 1);
										setDirty(false);
										onDirty(false);
										setLocalError(undefined);
									}
								} catch {
									setLocalError({ kind: "error", message: "Revisi staf tidak dapat dimuat ulang. Isian tetap dipertahankan." });
								} finally {
									setReloading(false);
								}
							}}
						>
							Muat ulang revisi staf
						</Button>
					)}
				</>
			)}
		</div>
	);
}
