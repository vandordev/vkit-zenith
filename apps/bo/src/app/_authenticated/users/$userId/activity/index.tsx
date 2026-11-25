import { createFileRoute } from "@tanstack/react-router";
import { userActivitySearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { HistorySection } from "../../../../-components/history-section";
import { pageSize, sectionOf } from "../../../organizations/$organizationId/-adapters/org-adapter";

export const Route = createFileRoute("/_authenticated/users/$userId/activity/")({ validateSearch: userActivitySearch, component: Activity });
function Activity() {
	const { userId } = Route.useParams();
	const { q, page, returnTo } = Route.useSearch();
	const navigate = Route.useNavigate();
	const list = usePrivateQuery(
		["customer-user-history", userId, q, page],
		async (t, signal) =>
			(await sdk.boV1ListUserHistory({ client: t.client, signal, throwOnError: true, path: { userId }, query: { page, pageSize, q: q || undefined } })).data,
	);
	return (
		<HistorySection
			description="Siapa atau apa yang mengubah akun pelanggan ini, beserta alasannya."
			q={q}
			entries={list.data?.items ?? []}
			pagination={list.data?.pagination}
			loading={list.isFetching}
			navigationKey={JSON.stringify([userId, page, returnTo])}
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
