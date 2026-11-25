import { z } from "zod";

// URL state of a settings list: text, status and page. Values that fail validation fall back to
// "unset" instead of breaking the page; the router parses `q=123` as a number, hence coerce.
export const settingsSearch = <const S extends readonly [string, ...string[]]>(statuses: S) =>
	z.object({
		q: z.coerce.string().max(100).optional().catch(undefined),
		status: z.enum(statuses).optional().catch(undefined),
		page: z.coerce.number().int().min(1).max(1000000).optional().catch(undefined),
	});

export const apiKeyStatuses = ["active", "revoked", "expired"] as const;
export const apiKeySearch = settingsSearch(apiKeyStatuses);

/** The filter's raw select value narrowed to the API's enum for this menu (anything else means "no status filter"). */
export const pickStatus = <const S extends readonly string[]>(values: S, value: string): S[number] | undefined =>
	values.find((item) => item === value);

/**
 * A dummy-written record joins the list only when the server's own filter would return it: the status matches, and the text (lower-cased,
 * as the handlers do) is contained in at least one of the fields the list handler searches, each checked on its own.
 */
export const matchesSettingsFilter = (
	search: { q?: string; status?: string },
	item: { status: string },
	fields: readonly (string | undefined)[],
) => {
	const q = (search.q ?? "").toLowerCase();
	return (
		(search.status === undefined || search.status === item.status) &&
		(q === "" || fields.some((field) => (field ?? "").toLowerCase().includes(q)))
	);
};
/** Exactly the fields of the list handlers (api_keys_list.go). */
export const searchFields = {
	apiKey: (item: { name: string; prefix: string }) => [item.name, item.prefix],
};
