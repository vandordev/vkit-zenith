import { useQueryClient } from "@tanstack/react-query";
import { errorMessage, type BOTransport } from "@bo/lib/api/client";
import { useProtected } from "@bo/lib/api/session";
import type { SectionState } from "../../../../-components/org-section";

/** Page size of every organization list; the page itself is validated URL search. */
export const pageSize = 25;

/** A 503 (or any failure) from the dummy API maps to the error state, never to empty. */
export function sectionOf(query: { isPending: boolean; error: unknown }): { state: SectionState; error?: string } {
	if (query.error) return { state: "error", error: errorMessage(query.error) };
	return { state: query.isPending ? "pending" : "ready" };
}

export function detailSectionOf<T>(query: { isPending: boolean; error: unknown }, value: T | undefined): { state: SectionState; error?: string } {
	if (query.error) return { state: "error", error: errorMessage(query.error) };
	if (query.isPending) return { state: "pending" };
	return value === undefined ? { state: "empty" } : { state: "ready" };
}

/**
 * Dummy adjustments: CSRF-protected write, then invalidate the affected list. The dummy API
 * persists nothing, so callers also overlay the returned result locally (keyed by ID).
 */
export function useAdjust() {
	const { transport, staff, epoch } = useProtected();
	const queries = useQueryClient();
	return async function adjust<T>(resource: string, organizationId: string, call: (transport: BOTransport) => Promise<T>): Promise<T> {
		const result = await call(transport);
		void queries.invalidateQueries({ queryKey: ["private", staff.id, epoch, resource, organizationId] });
		return result;
	};
}
