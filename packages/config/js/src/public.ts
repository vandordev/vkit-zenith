import { z } from "zod";

export const apiBaseUrlSchema = z
	.string()
	.url()
	.refine((value) => {
		try {
			const url = new URL(value);
			return (
				["http:", "https:"].includes(url.protocol) && !url.username && !url.password && !url.hash && !url.search && url.pathname === "/"
			);
		} catch {
			return false;
		}
	}, "expected an HTTP(S) origin without credentials");
export const publicBOConfigSchema = z.strictObject({ boApiBaseUrl: apiBaseUrlSchema });
export type PublicBOConfig = z.output<typeof publicBOConfigSchema>;
export const publicWebConfigSchema = z.strictObject({ apiBaseUrl: apiBaseUrlSchema });
export type PublicWebConfig = z.output<typeof publicWebConfigSchema>;
