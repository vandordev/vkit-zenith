import { ZenithLoader } from "@repo/ui/zenith-loader";
import { AuthLayout } from "./auth-layout";
import { paths } from "../../lib/paths";

export function RoutePending() {
	return (
		<main className="bg-background text-foreground">
			<ZenithLoader mode="screen" />
		</main>
	);
}
export function RouteError() {
	return (
		<AuthLayout
			title="Halaman tidak tersedia"
			description="Muat ulang halaman atau kembali ke beranda."
			footer={<a href={paths.home}>Beranda</a>}
		>
			{null}
		</AuthLayout>
	);
}
export function RouteNotFound() {
	return (
		<AuthLayout title="Halaman tidak ditemukan" footer={<a href={paths.home}>Beranda</a>}>
			{null}
		</AuthLayout>
	);
}
