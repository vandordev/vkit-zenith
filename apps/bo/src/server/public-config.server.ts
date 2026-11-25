import "@tanstack/react-start/server-only";
import { setResponseHeader } from "@tanstack/react-start/server";
import { getPublicBOConfig } from "@repo/config/server";
import { publicBOConfigSchema } from "@repo/config/public";
export function readBOPublicConfig() {
	setResponseHeader("Cache-Control", "no-store");
	try {
		const { boApiBaseUrl } = getPublicBOConfig();
		return publicBOConfigSchema.parse({ boApiBaseUrl });
	} catch {
		throw new Error("Konfigurasi backoffice tidak tersedia.");
	}
}
