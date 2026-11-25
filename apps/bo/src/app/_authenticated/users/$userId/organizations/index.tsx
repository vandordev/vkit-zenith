import { createFileRoute, Link } from "@tanstack/react-router";
import { customerOrigin, userOrganizationsSearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { pageSize, sectionOf } from "../../../organizations/$organizationId/-adapters/org-adapter";
import { UserOrganizations } from "./-components/user-organizations";

export const Route = createFileRoute("/_authenticated/users/$userId/organizations/")({
	validateSearch: userOrganizationsSearch,
	component: Organizations,
});
function Organizations() {
	const { userId } = Route.useParams();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const memberships = usePrivateQuery(
		["customer-user-memberships", userId, search.page],
		async (t, signal) =>
			(
				await sdk.boV1ListUserMemberships({
					client: t.client,
					signal,
					throwOnError: true,
					path: { id: userId },
					query: { page: search.page, pageSize },
				})
			).data,
	);
	// The member workspace returns to the original list (or to this tab); a return chain is never nested.
	const origin = customerOrigin(`/users/${userId}/organizations`, search);
	return (
		<UserOrganizations
			memberships={memberships.data}
			loading={memberships.isFetching}
			{...sectionOf(memberships)}
			onPage={(page) => {
				void navigate({ search: (previous) => ({ ...previous, page }) });
			}}
			onRetry={() => {
				void memberships.refetch();
			}}
			organizationLink={(row) => (
				<Link
					to="/organizations/$organizationId/members"
					params={{ organizationId: row.organizationId }}
					search={{ q: "", status: "", ownership: "", page: 1, selected: row.id, returnTo: origin }}
				/>
			)}
		/>
	);
}
