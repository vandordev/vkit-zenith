const invalidRange = () => new Error("Rentang tanggal tidak valid.");

function localDate(value: string): Date {
	if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) throw invalidRange();
	const [year, month, day] = value.split("-").map(Number);
	if (year === undefined || month === undefined || day === undefined || year < 1) throw invalidRange();
	const date = new Date(0);
	date.setFullYear(year, month - 1, day);
	date.setHours(0, 0, 0, 0);
	if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) throw invalidRange();
	return date;
}

// Browser-local calendar boundaries, never a fixed 24-hour duration.
export function activityRange(from?: string, to?: string): { from?: string; until?: string } {
	const start = from ? localDate(from) : undefined;
	const end = to ? localDate(to) : undefined;
	if (start && end && start > end) throw invalidRange();
	if (end) end.setDate(end.getDate() + 1);
	return { ...(start ? { from: start.toISOString() } : {}), ...(end ? { until: end.toISOString() } : {}) };
}
