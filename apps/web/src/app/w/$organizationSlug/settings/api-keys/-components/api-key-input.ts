import type { ApiKeyCreateInputBody } from "../../../../../../lib/api/generated/types.gen";
import { formProblemMessage } from "../../../-components/field-problem";
import { FormProblem } from "../../../-components/form-input";

export const apiKeyNameMax = 120;
/** `date` is the form's yyyy-mm-dd value: the key expires at the end of that local day. */
export type ApiKeyDraft = { name: string; date: string };
export const emptyApiKeyDraft: ApiKeyDraft = { name: "", date: "" };

const two = (value: number) => String(Math.abs(value)).padStart(2, "0");
/** An RFC 3339 instant that carries the local UTC offset (the API accepts offsets), second precision. */
export function toRfc3339Offset(date: Date): string {
	const offset = -date.getTimezoneOffset();
	const sign = offset < 0 ? "-" : "+";
	return (
		`${date.getFullYear()}-${two(date.getMonth() + 1)}-${two(date.getDate())}T${two(date.getHours())}:${two(date.getMinutes())}:${two(date.getSeconds())}` +
		`${sign}${two(Math.trunc(Math.abs(offset) / 60))}:${two(Math.abs(offset) % 60)}`
	);
}
/** End of the chosen local day, or undefined when the text is not a real calendar date. */
export function endOfDay(date: string): Date | undefined {
	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date);
	if (!match) return undefined;
	const [year, month, day] = [Number(match[1]), Number(match[2]), Number(match[3])];
	const value = new Date(year, month - 1, day, 23, 59, 59);
	return value.getFullYear() === year && value.getMonth() === month - 1 && value.getDate() === day ? value : undefined;
}

export function apiKeyProblems(draft: ApiKeyDraft, now: Date): Partial<Record<keyof ApiKeyDraft, string>> {
	const name = draft.name.trim();
	const expiry = draft.date === "" ? undefined : endOfDay(draft.date);
	return {
		...((name.length === 0 || name.length > apiKeyNameMax) && { name: `Nama wajib diisi, maksimal ${apiKeyNameMax} karakter.` }),
		...(draft.date !== "" && (!expiry || expiry <= now) && { date: "Pilih tanggal kedaluwarsa yang belum lewat, atau kosongkan." }),
	};
}
export function createApiKeyBody(draft: ApiKeyDraft, now: Date): ApiKeyCreateInputBody {
	if (Object.keys(apiKeyProblems(draft, now)).length > 0) throw new FormProblem(formProblemMessage);
	const expiry = draft.date === "" ? undefined : endOfDay(draft.date);
	return { name: draft.name.trim(), ...(expiry && { expiresAt: toRfc3339Offset(expiry) }) };
}
