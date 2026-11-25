import type { AuthTransport } from "./api/client";
import type { V1Page } from "./api/generated/types.gen";

// Pure organization helpers shared by route adapters and presentation (stories, fixtures): no generated SDK, runtime client or config imports,
// only types. lib/api/organization.ts re-exports them for route callers; presentation imports this module directly.

/** Every organization cache entry is scoped to identity epoch and organization. */
export const orgQueryKey = (t: AuthTransport, organizationId: string, ...parts: unknown[]): unknown[] => [
	"organization",
	t.epoch(),
	organizationId,
	...parts,
];

/** Fixed copy per status; server `detail` text is never shown for organization operations. */
export function orgErrorMessage(status: number): string {
	switch (status) {
		case 401:
			return "Sesi Anda berakhir. Masuk kembali untuk melanjutkan.";
		case 403:
			return "Peran Anda tidak memiliki izin untuk tindakan ini.";
		case 404:
			return "Data tidak ditemukan.";
		case 409:
			return "Tindakan tidak dapat dilakukan pada kondisi saat ini.";
		case 422:
			return "Periksa kembali isian Anda.";
		case 503:
			return "Layanan sedang tidak tersedia. Coba lagi sebentar lagi.";
		default:
			return "Terjadi kendala. Coba lagi sebentar lagi.";
	}
}

/** The HTTP status carried by a failed call (AuthError has one); anything else is unknown. Structural, so presentation never needs the error class. */
export const errorStatus = (error: unknown): number | undefined =>
	typeof error === "object" && error !== null && "status" in error && typeof error.status === "number" ? error.status : undefined;
/** The only text a dialog may show for a failed write: fixed copy by status, never server detail. */
export const orgFailureMessage = (error: unknown) => orgErrorMessage(errorStatus(error) ?? 0);

// Reference data (pickers, human-readable names) needs every record, not the first page, and must never present a partial list as complete.
/** More records than the page guard allows: a permanent capacity limit, retrying will not help. */
export class ListTooLarge extends Error {}
/** The server's paging contradicted itself (an empty page before the reported total): possibly transient, so an ordinary failure with retry. */
export class PagingStalled extends Error {}
/** How a reference list failed, for the copy and whether to offer retry. */
export const listErrorKind = (error: unknown): "failed" | "tooLarge" => (error instanceof ListTooLarge ? "tooLarge" : "failed");
// Fail-closed bound on pages, not a silent cap: past it the load errors with ListTooLarge instead of returning what it has.
const maxPages = 200;
export async function allPages<T>(fetchPage: (page: number) => Promise<{ items: T[] | null; pagination: V1Page }>): Promise<T[]> {
	const all: T[] = [];
	for (let page = 1; page <= maxPages; page++) {
		const result = await fetchPage(page);
		all.push(...(result.items ?? []));
		if (BigInt(result.pagination.total) <= BigInt(all.length)) return all;
		if (!result.items?.length) throw new PagingStalled("empty page before the reported total");
	}
	throw new ListTooLarge("more pages than the guard allows");
}
/**
 * placeholderData for paged lists: keep the previous page only while epoch and organization are unchanged.
 * Another organization or identity shows the pending state instead of the previous partition's rows.
 */
export const keepSameScope =
	(t: AuthTransport, organizationId: string) =>
	<T>(previous: T | undefined, previousQuery: { queryKey: readonly unknown[] } | undefined): T | undefined =>
		previousQuery && sameScope(previousQuery.queryKey, t.epoch(), organizationId) ? previous : undefined;
export const sameScope = (key: readonly unknown[], epoch: unknown, organizationId: string) =>
	key[0] === "organization" && key[1] === epoch && key[2] === organizationId;
