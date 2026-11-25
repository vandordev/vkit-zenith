import { createFileRoute } from "@tanstack/react-router";
import { organizationSectionSearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { pageSize, sectionOf } from "../-adapters/org-adapter";
import { OrganizationActivity } from "./-components/organization-activity";

export const Route = createFileRoute("/_authenticated/organizations/$organizationId/activity/")({
	validateSearch: organizationSectionSearch,
	component: Activity,
});
function Activity() {
	const { organizationId } = Route.useParams();
	const { page } = Route.useSearch();
	const navigate = Route.useNavigate();
	const list = usePrivateQuery(
		["org-history", organizationId, page],
		async (t, signal) =>
			(
				await sdk.boV1ListHistory({
					client: t.client,
					signal,
					throwOnError: true,
					path: { organizationId },
					query: { page, pageSize },
				})
			).data,
	);
	const entries = list.data?.items ?? [];
	const section = sectionOf(list);
	return (
		<OrganizationActivity
			entries={entries}
			state={section.state}
			error={section.error}
			pagination={list.data?.pagination}
			loading={list.isFetching}
			onPage={(next) => {
				void navigate({ search: (previous) => ({ ...previous, page: next }) });
			}}
			onRetry={() => {
				void list.refetch();
			}}
		/>
	);
}
