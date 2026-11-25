import { createFileRoute } from "@tanstack/react-router";
import { userSessionsSearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { pageSize, sectionOf } from "../../../organizations/$organizationId/-adapters/org-adapter";
import { UserSessions } from "./-components/user-sessions";

export const Route = createFileRoute("/_authenticated/users/$userId/sessions/")({ validateSearch: userSessionsSearch, component: Sessions });
function Sessions() {
	const { userId } = Route.useParams();
	const { page } = Route.useSearch();
	const navigate = Route.useNavigate();
	const list = usePrivateQuery(
		["customer-user-sessions", userId, page],
		async (t, signal) =>
			(await sdk.boV1ListUserSessions({ client: t.client, signal, throwOnError: true, path: { userId }, query: { page, pageSize } })).data,
	);
	return (
		<UserSessions
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
