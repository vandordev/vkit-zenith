import { useState } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useOrg } from "../../-components/org-context";
import { writeAccess } from "../-components/settings-access";
import { ProfileView } from "./-components/profile-view";
import { ApiGate } from "../../../../../lib/api/provider";
import type { AuthTransport } from "../../../../../lib/api/client";
import type { LogoResult } from "../../../../../lib/api/generated/types.gen";
import { errorStatus, getProfile, orgErrorMessage, orgFailureMessage, orgQueryKey, uploadLogo } from "../../../../../lib/api/organization";

export const Route = createFileRoute("/w/$organizationSlug/settings/profile/")({
	component: () => <ApiGate>{({ transport }) => <Entry t={transport} />}</ApiGate>,
});

// Keyed by organization (not mode: this page is not mode-scoped): a simulated logo never leaks into another organization.
function Entry({ t }: { t: AuthTransport }) {
	const { organization } = useOrg();
	return <Page key={organization.id} t={t} />;
}

function Page({ t }: { t: AuthTransport }) {
	const { organization, can } = useOrg();
	const [simulatedLogo, setSimulatedLogo] = useState<LogoResult>();
	// Not mode-scoped: the key carries no mode.
	const query = useQuery({
		queryKey: orgQueryKey(t, organization.id, "none", "profile"),
		queryFn: () => getProfile(t, organization.id),
	});
	return (
		<ProfileView
			state={query.isError ? "error" : query.data ? "ready" : "pending"}
			profile={query.data}
			access={writeAccess(can("providers.manage"), organization.status)}
			simulatedLogo={simulatedLogo}
			errorMessage={query.isError ? orgErrorMessage(errorStatus(query.error) ?? 0) : undefined}
			describeError={orgFailureMessage}
			onRetry={() => void query.refetch()}
			onUploadLogo={(body) => uploadLogo(t, organization.id, body)}
			onLogoResult={setSimulatedLogo}
		/>
	);
}
