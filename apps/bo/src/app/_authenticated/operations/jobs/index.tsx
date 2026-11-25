import { createFileRoute } from "@tanstack/react-router";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { collectionState, listHandlers, pageSize } from "../../-adapters/global-list";
import { jobsSearch } from "../../-adapters/global-search";
import { JobCollection } from "./-components/job-collection";

export const Route = createFileRoute("/_authenticated/operations/jobs/")({ validateSearch: jobsSearch, component: Jobs });
function Jobs() {
	const search = Route.useSearch();
	const list = usePrivateQuery(
		["ops-jobs", search.q, search.status, search.stateGroup, search.page],
		async (t, signal) =>
			(await sdk.boV1ListJobs({ client: t.client, signal, throwOnError: true, query: { page: search.page, pageSize, q: search.q, status: search.status, stateGroup: search.stateGroup } })).data,
	);
	return (
		<JobCollection
			{...listHandlers(Route.useNavigate() as never)}
			q={search.q}
			navigationKey={JSON.stringify(search)}
			status={search.status}
			stateGroup={search.stateGroup}
			items={list.data?.items ?? []}
			pagination={list.data?.pagination}
			{...collectionState(list)}
		/>
	);
}
