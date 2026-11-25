import { EditForm, type Feedback } from "../../../-components/common";
import type { V1UserCreateInputBody } from "@bo/lib/api/generated/types.gen";
export function UserCreate({
	pending,
	feedback,
	onDirty,
	onSubmit,
}: {
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onSubmit: (body: V1UserCreateInputBody) => void;
}) {
	return (
		<EditForm
			fields={[
				{ name: "displayName", label: "Nama pelanggan", maxLength: 200 },
				{ name: "email", label: "Email pelanggan", type: "email", maxLength: 254 },
			]}
			pending={pending}
			feedback={feedback}
			submitLabel="Buat pelanggan"
			onDirty={onDirty}
			onSubmit={(values) => onSubmit({ displayName: values.displayName ?? "", email: values.email ?? "" })}
		>
			<p className="max-w-[65ch] text-sm text-muted-foreground">
				Pelanggan baru berstatus aktif dengan email belum terverifikasi. Tindakan ini tidak membuat kata sandi atau akses masuk, dan tidak
				menambahkan keanggotaan organisasi.
			</p>
		</EditForm>
	);
}
