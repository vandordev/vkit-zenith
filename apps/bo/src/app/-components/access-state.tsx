import { ZenithLoader } from "@repo/ui/zenith-loader";
import { Page } from "./common";
import { ContentState } from "./content-state";

export function AccessState({ kind, onRetry }: { kind: "pending" | "config" | "denied" | "session"; onRetry?: () => void }) {
	if (kind === "pending") return <ZenithLoader mode="screen" />;
	return (
		<Page title={kind === "denied" ? "Akses ditolak" : kind === "session" ? "Sesi tidak dapat dimuat" : "Konfigurasi tidak tersedia"}>
			<ContentState
				kind="error"
				title={
					kind === "denied"
						? "Hanya Superadmin yang dapat mengelola staf."
						: kind === "session"
							? "Identitas staf belum dapat dimuat."
							: "Backoffice belum dapat terhubung."
				}
				description={
					kind === "denied" ? "Hubungi Superadmin jika akses diperlukan." : "Coba muat ulang atau hubungi operator jika masalah berlanjut."
				}
				onRetry={onRetry}
			/>
		</Page>
	);
}
