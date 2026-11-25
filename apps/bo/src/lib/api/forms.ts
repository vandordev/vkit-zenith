import { z } from "zod";
export function passwordValid(value: string): boolean {
	const count = [...value].length;
	return count >= 15 && count <= 128 && new TextEncoder().encode(value).length <= 512;
}
export const passwordSchema = z.string().refine(passwordValid, "Gunakan 15–128 karakter dan maksimal 512 byte UTF-8.");
export function formInput<T>(schema: z.ZodType<T>, value: unknown): T {
	const parsed = schema.safeParse(value);
	if (!parsed.success) throw new Error("Periksa isian wajib, format kunci, pilihan dan batas panjang. Data belum dikirim.");
	return parsed.data;
}
export function field(values: Record<string, string>, name: string): string {
	return values[name] ?? "";
}
