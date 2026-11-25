import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditForm, labels, type Feedback, type Field } from "../../../-components/common";

const options = (...values: string[]) => values.map((value) => ({ value, label: labels[value] ?? value }));
const staffFields: Field[] = [
	{ name: "email", label: "Email staf", type: "email", maxLength: 254 },
	{ name: "displayName", label: "Nama staf" },
	{ name: "role", label: "Peran staf", options: options("admin", "superadmin") },
	{ name: "status", label: "Status", options: options("active", "disabled") },
];
const confirmationField: Field = {
	name: "currentPassword",
	label: "Kata sandi Anda saat ini",
	type: "password",
	autoComplete: "current-password",
	help: "Konfirmasi hanya berlaku untuk perubahan ini.",
};
const newPasswordField: Field = {
	name: "password",
	label: "Kata sandi staf",
	type: "password",
	help: "15–128 karakter dan maksimal 512 byte UTF-8. Tidak ada kewajiban ganti pada login pertama.",
};

export function StaffEditor({
	initial,
	creating,
	replacing,
	pending,
	feedback,
	onSubmit,
	onDirty,
}: {
	initial: Record<string, string>;
	creating?: boolean;
	replacing?: boolean;
	pending?: boolean;
	feedback?: Feedback;
	onSubmit: (values: Record<string, string>) => void;
	onDirty: (dirty: boolean) => void;
}) {
	const [confirming, setConfirming] = useState(false);
	const [draft, setDraft] = useState<Record<string, string>>(initial);
	if (confirming || replacing)
		return (
			<div className="flex flex-col gap-5">
				<h2 className="font-medium">
					{replacing ? "Ganti kata sandi" : "Konfirmasi perubahan"}: {initial.displayName || draft.displayName || "staf baru"}
				</h2>
				<p className="text-sm leading-relaxed text-muted-foreground">
					Perubahan peran, status atau kata sandi mengakhiri semua sesi staf terkait. Setidaknya satu Superadmin aktif harus tetap tersedia.
				</p>
				<EditForm
					key="confirmation"
					fields={replacing || creating ? [newPasswordField, confirmationField] : [confirmationField]}
					pending={pending}
					feedback={feedback}
					submitLabel="Konfirmasi dan simpan"
					onSubmit={(values) => onSubmit({ ...draft, ...values })}
					onDirty={onDirty}
				/>
				{!replacing && (
					<Button variant="outline" disabled={pending} onClick={() => setConfirming(false)}>
						Kembali ke isian
					</Button>
				)}
			</div>
		);
	return (
		<EditForm
			key="metadata"
			initial={draft}
			fields={staffFields}
			pending={pending}
			feedback={feedback}
			submitLabel="Tinjau perubahan"
			onDirty={onDirty}
			onSubmit={(values) => {
				setDraft(values);
				setConfirming(true);
			}}
		/>
	);
}
