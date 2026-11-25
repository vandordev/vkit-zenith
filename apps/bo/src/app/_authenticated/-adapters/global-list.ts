import { useQueryClient } from "@tanstack/react-query";
import { errorMessage, type BOTransport } from "@bo/lib/api/client";
import { useProtected } from "@bo/lib/api/session";
import type { SectionState } from "../../-components/org-section";

export { pageSize } from "../organizations/$organizationId/-adapters/org-adapter";

type Search = Record<string, unknown>;
type Navigate = (options: { search: (previous: Search) => Search }) => unknown;

/** URL-state handlers every global list shares: a filter change restarts at page 1 and keeps the other filters. */
export function listHandlers(navigate: Navigate) {
	return {
		onFilter: (q: string, values: Record<string, string>) => void navigate({ search: (previous) => ({ ...previous, q, ...values, page: 1 }) }),
		onPage: (page: number) => void navigate({ search: (previous) => ({ ...previous, page }) }),
		onFilterOrganization: (organizationId: string) => void navigate({ search: (previous) => ({ ...previous, organizationId, page: 1 }) }),
		onClearOrganization: () => void navigate({ search: (previous) => ({ ...previous, organizationId: undefined, page: 1 }) }),
	};
}

/** Query lifecycle -> the collection's section state; a 503 is an error, never an empty list. */
export function collectionState(query: { isPending: boolean; error: unknown; isFetching: boolean; refetch: () => unknown }): {
	state: SectionState;
	error?: string;
	loading: boolean;
	onRetry: () => void;
} {
	return {
		state: query.error ? "error" : query.isPending ? "pending" : "ready",
		error: query.error ? errorMessage(query.error) : undefined,
		loading: query.isFetching,
		onRetry: () => void query.refetch(),
	};
}

/** Dummy operations action: CSRF-protected write, then refresh the resource's queries. The caller overlays the returned record. */
export function useOpsAction() {
	const { transport, staff, epoch } = useProtected();
	const queries = useQueryClient();
	return async function act<T>(resource: string, call: (transport: BOTransport) => Promise<T>): Promise<T> {
		const result = await call(transport);
		void queries.invalidateQueries({ queryKey: ["private", staff.id, epoch, resource] });
		return result;
	};
}
