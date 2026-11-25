import { exponentOf } from "@repo/ui/money";

/** A validation failure the user can fix; its message is fixed Indonesian copy and safe to show. */
export class FormProblem extends Error {}

/** "12.5" → digits scaled by `decimals` as a digit string, or undefined when not a plain decimal or more precise than `decimals`. */
function scaled(input: string, decimals: number): string | undefined {
	const [whole = "", fraction = "", ...extra] = input.trim().split(".");
	if (extra.length > 0 || !/^\d+$/.test(whole) || !/^\d*$/.test(fraction) || fraction.length > decimals) return undefined;
	return `${whole}${fraction.padEnd(decimals, "0")}`.replace(/^0+(?=\d)/, "");
}

/**
 * Typed major units ("125000", "12.5") to the API's integer minor-unit string, with string maths only (ADR-0005).
 * undefined when not exactly representable in the currency (too many decimals, unsupported currency, over 18 digits).
 */
export function toMinorUnits(input: string, currency: string): string | undefined {
	const exponent = exponentOf(currency);
	const minor = exponent === null ? undefined : scaled(input, exponent);
	return minor !== undefined && /^(0|[1-9]\d{0,17})$/.test(minor) ? minor : undefined;
}

/** Why an amount was refused: exponent 0 takes whole numbers, otherwise at most that many decimals. */
export function amountProblem(currency: string): string {
	const exponent = exponentOf(currency);
	return exponent === null
		? "Mata uang tidak didukung."
		: exponent === 0
			? "Masukkan bilangan bulat, tanpa pecahan."
			: `Masukkan nominal dengan maksimal ${exponent} desimal.`;
}

/** Percent with up to two decimals ("12.5") to basis points 1–10000, exactly; undefined when out of range or inexact. */
export function toBasisPoints(input: string): number | undefined {
	const points = scaled(input, 2);
	if (points === undefined || points.length > 5) return undefined;
	return Number(points) >= 1 && Number(points) <= 10000 ? Number(points) : undefined;
}

/** Runs a body builder; a FormProblem moves focus to the first marked field once the dialog has rendered its problems. */
export function buildBody<T>(build: () => T): T {
	try {
		return build();
	} catch (failure) {
		if (failure instanceof FormProblem)
			requestAnimationFrame(() =>
				requestAnimationFrame(() => document.querySelector<HTMLElement>('[role="dialog"] [aria-invalid="true"]')?.focus()),
			);
		throw failure;
	}
}
