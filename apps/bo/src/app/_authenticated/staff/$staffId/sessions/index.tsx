import { createFileRoute } from "@tanstack/react-router";
import { staffSectionSearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { pageSize, sectionOf } from "../../../organizations/$organizationId/-adapters/org-adapter";
import { StaffSessions } from "./-components/staff-sessions";

export const Route = createFileRoute("/_authenticated/staff/$staffId/sessions/")({ validateSearch: staffSectionSearch, component: Sessions });
function Sessions() {
	const { staffId } = Route.useParams();
	const { page } = Route.useSearch();
	const navigate = Route.useNavigate();
	const list = usePrivateQuery(
		["staff-sessions", staffId, page],
		async (t, signal) =>
			(await sdk.boV1ListStaffSessions({ client: t.client, signal, throwOnError: true, path: { staffId }, query: { page, pageSize } })).data,
	);
	return (
		<StaffSessions
			sessions={list.data?.items ?? []}
			pagination={list.data?.pagination}
			loading={list.isFetching}
			{...sectionOf(list)}
			onPage={(next) => {
				void navigate({ search: (previous) => ({ ...previous, page: next }) });
			}}
			onRetry={() => {
				void list.refetch();
			}}
		/>
	);
}
