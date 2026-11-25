import type { LogoUploadInputBody } from "../../../../../../lib/api/generated/types.gen";

// The dummy upload takes a content type and a size only; no file is read, sent or stored (real upload is M4).
export const logoContentTypes = ["image/png", "image/jpeg", "image/svg+xml"] as const;
export const logoMaxBytes = 1048576;
export type LogoChoice = { contentType: string; size: number };

const isLogoType = (value: string): value is LogoUploadInputBody["contentType"] => (logoContentTypes as readonly string[]).includes(value);
export function logoProblem(choice: LogoChoice): string | undefined {
	if (!isLogoType(choice.contentType)) return "Format logo harus PNG, JPEG atau SVG.";
	if (choice.size < 1) return "Berkas kosong.";
	if (choice.size > logoMaxBytes) return "Ukuran logo maksimal 1 MB.";
	return undefined;
}
/** The request body, or undefined when the choice is refused. */
export function logoBody(choice: LogoChoice): LogoUploadInputBody | undefined {
	return isLogoType(choice.contentType) && !logoProblem(choice) ? { contentType: choice.contentType, size: choice.size } : undefined;
}
