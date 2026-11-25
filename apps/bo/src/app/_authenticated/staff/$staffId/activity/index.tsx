import { createFileRoute } from "@tanstack/react-router";
import { staffActivitySearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { HistorySection } from "../../../../-components/history-section";
import { pageSize, sectionOf } from "../../../organizations/$organizationId/-adapters/org-adapter";

export const Route = createFileRoute("/_authenticated/staff/$staffId/activity/")({ validateSearch: staffActivitySearch, component: Activity });
function Activity() {
	const { staffId } = Route.useParams();
	const { q, page, returnTo } = Route.useSearch();
	const navigate = Route.useNavigate();
	// The `activity` prefix keeps the shell's existing cache removal for non-superadmin sessions.
	const list = usePrivateQuery(
		["activity-staff-history", staffId, q, page],
		async (t, signal) =>
			(await sdk.boV1ListStaffHistory({ client: t.client, signal, throwOnError: true, path: { staffId }, query: { page, pageSize, q: q || undefined } })).data,
	);
	return (
		<HistorySection
			description="Siapa atau apa yang mengubah akun staf ini, beserta alasannya."
			q={q}
			entries={list.data?.items ?? []}
			pagination={list.data?.pagination}
			loading={list.isFetching}
			navigationKey={JSON.stringify([staffId, page, returnTo])}
			{...sectionOf(list)}
			onSearch={(next) => {
				void navigate({ search: (previous) => ({ ...previous, q: next, page: 1 }) });
			}}
			onPage={(next) => {
				void navigate({ search: (previous) => ({ ...previous, page: next }) });
			}}
			onRetry={() => {
				void list.refetch();
			}}
		/>
	);
}
