import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useOrg } from "../../-components/org-context";
import { MembersList } from "./-components/members-list";
import { ApiGate } from "../../../../../lib/api/provider";
import type { AuthTransport } from "../../../../../lib/api/client";
import { errorStatus, listMembers, orgErrorMessage, orgQueryKey } from "../../../../../lib/api/organization";

export const Route = createFileRoute("/w/$organizationSlug/settings/members/")({
	component: () => <ApiGate>{({ transport }) => <Page t={transport} />}</ApiGate>,
});

function Page({ t }: { t: AuthTransport }) {
	const { organization } = useOrg();
	const query = useQuery({
		queryKey: orgQueryKey(t, organization.id, "members"),
		queryFn: () => listMembers(t, organization.id),
	});
	return (
		<MembersList
			state={query.isError ? "error" : query.data ? "ready" : "pending"}
			members={query.data}
			errorMessage={query.isError ? orgErrorMessage(errorStatus(query.error) ?? 0) : undefined}
			onRetry={() => void query.refetch()}
		/>
	);
}
