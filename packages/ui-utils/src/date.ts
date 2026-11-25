export type DateInput = string | number | Date | null | undefined;
export type DateFormatOptions = {
	locale?: string | string[];
	timeZone?: string;
	format?: Intl.DateTimeFormatOptions;
	fallback?: string;
};

/** Timestamp strings must include a timezone. Date-only values are not instants. */
export function toDate(value: DateInput): Date | null {
	if (value == null || value === "") return null;
	if (typeof value === "string") {
		const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.exec(value);
		if (!match) return null;
		const [, year, month, day, hour, minute, second] = match;
		const calendar = new Date(`${year}-${month}-${day}T00:00:00Z`);
		if (
			calendar.getUTCFullYear() !== Number(year) ||
			calendar.getUTCMonth() + 1 !== Number(month) ||
			calendar.getUTCDate() !== Number(day) ||
			Number(hour) > 23 ||
			Number(minute) > 59 ||
			Number(second) > 59
		)
			return null;
	}
	const date = new Date(value instanceof Date ? value.getTime() : value);
	return Number.isFinite(date.getTime()) ? date : null;
}

/** Without timeZone, Intl uses the current runtime's zone; use the React adapter for SSR. */
export function formatDate(value: DateInput, options: DateFormatOptions = {}): string {
	const date = toDate(value);
	if (!date) return options.fallback ?? "—";
	return new Intl.DateTimeFormat(options.locale ?? "id-ID", {
		...(options.format ?? { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit", hourCycle: "h23" }),
		...(options.timeZone === undefined ? {} : { timeZone: options.timeZone }),
	}).format(date);
}
