import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditForm, type Feedback } from "../../../-components/common";
import { ContentState } from "../../../-components/content-state";
import type { V1OrganizationGetBody, V1OrganizationUpdateInputBody } from "@bo/lib/api/generated/types.gen";

export function OrganizationProfile({
	record,
	pending,
	feedback,
	refreshError,
	onDirty,
	onSubmit,
	onReload,
}: {
	record: V1OrganizationGetBody;
	pending?: boolean;
	feedback?: Feedback;
	refreshError?: string;
	onDirty: (dirty: boolean) => void;
	onSubmit: (body: V1OrganizationUpdateInputBody) => void;
	onReload: () => Promise<V1OrganizationGetBody | undefined>;
}) {
	const [snapshot, setSnapshot] = useState(record);
	const [version, setVersion] = useState(0);
	const [dirty, setDirty] = useState(false);
	const [reloading, setReloading] = useState(false);
	const [reloadFeedback, setReloadFeedback] = useState<Feedback>();
	const reload = async () => {
		if (pending || reloading || (dirty && !window.confirm("Buang perubahan yang belum disimpan dan muat ulang organisasi?"))) return;
		setReloading(true);
		try {
			const fresh = await onReload();
			if (!fresh || fresh.id !== snapshot.id) {
				setReloadFeedback({ kind: "error", message: "Organisasi tidak dapat dimuat. Isian tetap dipertahankan." });
				return;
			}
			setSnapshot(fresh);
			setVersion((n) => n + 1);
			setDirty(false);
			onDirty(false);
			setReloadFeedback(undefined);
		} catch {
			setReloadFeedback({ kind: "error", message: "Organisasi tidak dapat dimuat. Isian tetap dipertahankan." });
		} finally {
			setReloading(false);
		}
	};
	return (
		<div className="space-y-5">
			{refreshError && <ContentState kind="error" title={refreshError} description="Isian dan revisi awal tetap dipertahankan." />}
			<EditForm
				key={version}
				initial={{ name: snapshot.name, slug: snapshot.slug }}
				fields={[
					{ name: "name", label: "Nama organisasi", maxLength: 200 },
					{ name: "slug", label: "Slug organisasi", maxLength: 64 },
				]}
				pending={pending || reloading}
				feedback={reloadFeedback ?? feedback}
				onDirty={(next) => {
					setDirty(next);
					onDirty(next);
				}}
				onSubmit={(values) => onSubmit({ name: values.name ?? "", slug: values.slug ?? "", revision: snapshot.revision })}
			/>
			{feedback?.kind === "error" && (
				<Button variant="outline" disabled={pending || reloading} onClick={() => void reload()}>
					Muat ulang revisi organisasi
				</Button>
			)}
			<p className="text-xs text-muted-foreground">Perubahan profil tidak mengubah status, anggota atau kepemilikan organisasi.</p>
		</div>
	);
}
