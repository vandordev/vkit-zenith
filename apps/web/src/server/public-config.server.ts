import "@tanstack/react-start/server-only";
import { setResponseHeader } from "@tanstack/react-start/server";
import { getPublicWebConfig } from "@repo/config/server";
import { publicWebConfigSchema } from "@repo/config/public";
export function readPublicConfig() {
	setResponseHeader("Cache-Control", "no-store");
	try {
		const { apiBaseUrl } = getPublicWebConfig();
		return publicWebConfigSchema.parse({ apiBaseUrl });
	} catch {
		throw new Error("Konfigurasi Auth tidak tersedia.");
	}
}
