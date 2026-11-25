import type { BoV1GetMeResponse } from "../../../../lib/api/generated/types.gen";
import { EditForm, Page, type Feedback, type Field } from "../../../-components/common";

const passwordFields: Field[] = [
	{ name: "currentPassword", label: "Kata sandi saat ini", type: "password", autoComplete: "current-password" },
	{
		name: "newPassword",
		label: "Kata sandi baru",
		type: "password",
		help: "15–128 karakter, maksimal 512 byte UTF-8. Tidak boleh sama dengan kata sandi saat ini.",
	},
	{ name: "confirmPassword", label: "Ulangi kata sandi baru", type: "password" },
];

export function AccountPage({
	staff,
	pending,
	feedback,
	onSubmit,
}: {
	staff: BoV1GetMeResponse;
	pending?: boolean;
	feedback?: Feedback;
	onSubmit: (values: Record<string, string>) => void;
}) {
	return (
		<Page title="Akun Saya" description="Perubahan kata sandi mengakhiri semua sesi Anda, termasuk sesi ini.">
			<dl className="grid max-w-xl grid-cols-[auto_minmax(0,1fr)] gap-x-8 gap-y-3 text-sm">
				<dt>Nama</dt>
				<dd className="break-words">{staff.displayName}</dd>
				<dt>Email</dt>
				<dd className="break-all">{staff.email}</dd>
				<dt>Peran</dt>
				<dd>{staff.role === "superadmin" ? "Superadmin" : "Admin"}</dd>
			</dl>
			<section className="max-w-lg border-t border-border pt-6">
				<h2 className="mb-5 text-lg font-semibold">Ganti kata sandi</h2>
				<EditForm
					fields={passwordFields}
					onSubmit={onSubmit}
					pending={pending}
					feedback={feedback}
					submitLabel="Ganti kata sandi dan keluar"
				/>
			</section>
		</Page>
	);
}
