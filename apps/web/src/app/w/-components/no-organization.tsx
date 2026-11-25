import { Button, buttonVariants } from "@repo/ui/button";
import { AuthLayout } from "../../-components/auth-layout";
import type { FormStatus } from "../../-components/form-status";
import { StatusNotice } from "../../-components/status-notice";

// Placeholder contact mirroring apps/landing (_components/landing-content.ts); web cannot import landing source.
const contactDisplay = "hello@example.com";
const contactHref = `mailto:${contactDisplay}?subject=${encodeURIComponent("Organisasi saya belum aktif")}`;

/** 0 active memberships: staff activate organizations, so the way forward is contact and logout, never a dead end. */
export function NoOrganization({ displayName, logout, onLogout }: { displayName: string; logout: FormStatus; onLogout: () => void }) {
	const pending = logout.kind === "pending";
	return (
		<AuthLayout
			title="Organisasi Anda belum aktif"
			description={`Halo ${displayName}. Organisasi didaftarkan dan diaktifkan oleh tim Zenith. Hubungi kami agar akun Anda ditautkan ke organisasi.`}
		>
			<div className="flex flex-col gap-3">
				<a href={contactHref} className={buttonVariants()}>
					Hubungi tim Zenith ({contactDisplay})
				</a>
				<StatusNotice status={logout} />
				<Button type="button" variant="outline" isLoading={pending} onClick={() => !pending && onLogout()}>
					Keluar
				</Button>
			</div>
		</AuthLayout>
	);
}
