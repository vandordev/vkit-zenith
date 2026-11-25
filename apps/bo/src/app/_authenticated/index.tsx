import { createFileRoute, Link } from "@tanstack/react-router";
import { HomeOverview } from "./-components/home-counters";
import { RecentActivity } from "./-components/recent-activity";
import { collectionState } from "./-adapters/global-list";
import { usePrivateQuery, useProtected } from "../../lib/api/session";
import { boV1GetHomeCounters, boV1GetRecentActivity } from "../../lib/api/generated/sdk.gen";
import { errorMessage } from "../../lib/api/client";
export const Route = createFileRoute("/_authenticated/")({ component: Home });
function Home() {
	const { staff, epoch } = useProtected();
	const counters = usePrivateQuery(
		["home-counters"],
		async (transport, signal) => (await boV1GetHomeCounters({ client: transport.client, throwOnError: true, signal })).data,
	);
	const { state, error, onRetry } = collectionState(counters);
	return (
		<HomeOverview counters={counters.data} state={counters.data ? "ready" : state} error={error} onRetry={onRetry}>
			{/* The activity log is superadmin-only; the key restarts it with the identity. */}
			{staff.role === "superadmin" && <RecentAdapter key={`${staff.id}:${epoch}`} />}
		</HomeOverview>
	);
}
function RecentAdapter() {
	const activity = usePrivateQuery(
		["activity-recent"],
		async (transport, signal) => (await boV1GetRecentActivity({ client: transport.client, throwOnError: true, signal })).data,
	);
	return (
		<RecentActivity
			{...{
				data: activity.data,
				loading: activity.isPending || activity.isFetching,
				error: activity.error ? errorMessage(activity.error) : undefined,
				onRetry: () => {
					void activity.refetch();
				},
				renderLink: (row) => (
					<Link to="/activity" search={{ selected: row.id, page: 1, invalid: false }} className="underline">
						Periksa aktivitas
					</Link>
				),
				renderAll: () => (
					<Link to="/activity" search={{ page: 1, invalid: false }} className="text-sm underline">
						Lihat log aktivitas
					</Link>
				),
			}}
		/>
	);
}
