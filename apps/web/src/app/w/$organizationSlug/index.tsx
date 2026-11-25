import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useOrg } from "./-components/org-context";
import { OrgOverview } from "./-components/org-overview";
import { ApiGate } from "../../../lib/api/provider";
import type { AuthTransport } from "../../../lib/api/client";
import { errorStatus, listMembers, orgErrorMessage, orgQueryKey } from "../../../lib/api/organization";

export const Route = createFileRoute("/w/$organizationSlug/")({
	component: () => <ApiGate>{({ transport }) => <Page t={transport} />}</ApiGate>,
});

function Page({ t }: { t: AuthTransport }) {
	const { organization } = useOrg();
	// Same key as the members page, so the list is fetched once and shared.
	const members = useQuery({
		queryKey: orgQueryKey(t, organization.id, "members"),
		queryFn: () => listMembers(t, organization.id),
	});
	return (
		<OrgOverview
			organization={organization}
			memberCount={members.data?.length}
			membersState={members.isError ? "error" : members.data ? "ready" : "pending"}
			membersError={members.isError ? orgErrorMessage(errorStatus(members.error) ?? 0) : undefined}
			onRetryMembers={() => void members.refetch()}
		/>
	);
}
