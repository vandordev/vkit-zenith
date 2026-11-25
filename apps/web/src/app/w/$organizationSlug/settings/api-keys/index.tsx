import { useState } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useOrg } from "../../-components/org-context";
import { listNavigationKey } from "../../-components/format";
import { applySimulation, createdIds, noSimulation, simulateChanged, simulateCreated, type Simulation } from "../../-components/simulated";
import { writeAccess } from "../-components/settings-access";
import { apiKeyStatuses, apiKeySearch, matchesSettingsFilter, pickStatus, searchFields } from "../-components/settings-search";
import { ApiKeysList } from "./-components/api-keys-list";
import { ApiGate } from "../../../../../lib/api/provider";
import type { AuthTransport } from "../../../../../lib/api/client";
import type { ApiKey } from "../../../../../lib/api/generated/types.gen";
import {
	createApiKey,
	errorStatus,
	keepSameScope,
	listApiKeys,
	orgErrorMessage,
	orgFailureMessage,
	orgQueryKey,
	revokeApiKey,
} from "../../../../../lib/api/organization";

export const Route = createFileRoute("/w/$organizationSlug/settings/api-keys/")({
	validateSearch: apiKeySearch,
	component: () => <ApiGate>{({ transport }) => <Entry t={transport} />}</ApiGate>,
});

// Keyed by organization: simulated results, and any dialog still holding a one-time secret, are dropped on a scope change.
function Entry({ t }: { t: AuthTransport }) {
	const { organization } = useOrg();
	return <Page key={organization.id} t={t} />;
}

function Page({ t }: { t: AuthTransport }) {
	const { organization, can } = useOrg();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const page = search.page ?? 1;
	const [simulation, setSimulation] = useState<Simulation<ApiKey>>(noSimulation);
	const query = useQuery({
		queryKey: orgQueryKey(t, organization.id, "api-keys", search.q ?? "", search.status ?? "", page),
		queryFn: () => listApiKeys(t, organization.id, { q: search.q, status: search.status, page }),
		placeholderData: keepSameScope(t, organization.id),
	});
	const items = applySimulation(
		query.data?.items ?? [],
		simulation,
		page,
		(key) => matchesSettingsFilter(search, key, searchFields.apiKey(key)),
		search.status !== undefined,
	);
	return (
		<ApiKeysList
			state={query.isError ? "error" : query.data ? "ready" : "pending"}
			keys={items}
			pagination={query.data?.pagination}
			filters={{ q: search.q ?? "", status: search.status ?? "" }}
			navigationKey={listNavigationKey(search)}
			access={writeAccess(can("developers.manage"), organization.status)}
			simulatedIds={createdIds(simulation)}
			simulated={simulation.created.length > 0 || Object.keys(simulation.changed).length > 0}
			loading={query.isPlaceholderData}
			errorMessage={query.isError ? orgErrorMessage(errorStatus(query.error) ?? 0) : undefined}
			describeError={orgFailureMessage}
			onFilter={({ q, status }) =>
				void navigate({
					search: (previous) => ({ ...previous, q: q || undefined, status: pickStatus(apiKeyStatuses, status), page: undefined }),
				})
			}
			onPage={(next) => void navigate({ search: (previous) => ({ ...previous, page: next > 1 ? next : undefined }) })}
			onRetry={() => void query.refetch()}
			// Direct SDK call, no useMutation or cache: the secret travels from this promise to the dialog's local state only.
			onCreate={(body) => createApiKey(t, organization.id, body)}
			onRevoke={(keyId) => revokeApiKey(t, organization.id, keyId)}
			onResult={(key, kind) =>
				setSimulation((current) => (kind === "created" ? simulateCreated(current, key) : simulateChanged(current, key)))
			}
		/>
	);
}
