/** Date without time for list cells. */
export const dateOnly: Intl.DateTimeFormatOptions = { day: "numeric", month: "short", year: "numeric" };

export const sizeText = (bytes: number) =>
	bytes >= 1024 * 1024 ? `${(bytes / 1024 / 1024).toFixed(1)} MB` : `${Math.max(1, Math.round(bytes / 1024))} KB`;

/** Search state that should move focus/scroll when it changes (not the typed query). */
export const listNavigationKey = (search: { q?: string; page?: number; status?: string }) =>
	JSON.stringify({ page: search.page ?? 1, status: search.status ?? "" });
