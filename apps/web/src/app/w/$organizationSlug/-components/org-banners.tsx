import { TriangleAlert } from "lucide-react";

export function SuspendedBanner() {
	return (
		<div
			data-slot="suspended-banner"
			role="status"
			className="flex min-w-0 items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive-text"
		>
			<TriangleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
			<span className="min-w-0 wrap-anywhere">
				<strong className="font-semibold">Organisasi ditangguhkan.</strong> Anda masih bisa melihat data, tetapi tindakan yang mengubah data
				dinonaktifkan. Hubungi tim Zenith.
			</span>
		</div>
	);
}

export function OrganizationBanners({ suspended }: { suspended: boolean }) {
	return suspended ? <SuspendedBanner /> : null;
}
