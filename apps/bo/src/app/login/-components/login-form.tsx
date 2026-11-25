import { brand } from "@repo/brand";
import { DisplayHeading } from "@repo/ui/display-heading";
import { ZenithMark } from "@repo/ui/zenith-mark";
import { EditForm, type Feedback } from "../../-components/common";

export function LoginForm({
	pending,
	feedback,
	onSubmit,
}: {
	pending?: boolean;
	feedback?: Feedback;
	onSubmit: (values: Record<string, string>) => void;
}) {
	return (
		<main className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-8 px-6 py-12">
			<header>
				<ZenithMark size={48} title={brand.name} />
				<p className="mt-5 text-sm text-muted-foreground">{brand.name} Backoffice</p>
				<DisplayHeading className="mt-2">Masuk untuk bekerja</DisplayHeading>
				<p className="mt-3 text-sm leading-relaxed text-muted-foreground">
					Khusus staf internal. Gunakan akun staf Anda, bukan akun pelanggan.
				</p>
			</header>
			<EditForm
				labelStyle="floating"
				pending={pending}
				feedback={feedback}
				submitLabel="Masuk"
				fields={[
					{ name: "email", label: "Email staf", type: "email", autoComplete: "username", maxLength: 254 },
					{ name: "password", label: "Kata sandi", type: "password", autoComplete: "current-password" },
				]}
				onSubmit={onSubmit}
			/>
		</main>
	);
}
