import { useState } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { organizationSectionSearch } from "@bo/lib/api/customer-navigation";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import type { ApiKey } from "@bo/lib/api/generated/types.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { AdjustmentDialog } from "../../../../-components/adjustment-dialog";
import { pageSize, sectionOf, useAdjust } from "../-adapters/org-adapter";
import { ApiKeyList, revokeApiKeyDialog, revokeApiKeyResult } from "./-components/api-key-list";

export const Route = createFileRoute("/_authenticated/organizations/$organizationId/api-keys/")({
	validateSearch: organizationSectionSearch,
	component: ApiKeys,
});
function ApiKeys() {
	const { organizationId } = Route.useParams();
	return <ApiKeysAdapter key={organizationId} organizationId={organizationId} />;
}
function ApiKeysAdapter({ organizationId }: { organizationId: string }) {
	const adjust = useAdjust();
	const { page } = Route.useSearch();
	const navigate = Route.useNavigate();
	const list = usePrivateQuery(
		["org-api-keys", organizationId, page],
		async (t, signal) =>
			(
				await sdk.boV1ListApiKeys({
					client: t.client,
					signal,
					throwOnError: true,
					path: { organizationId },
					query: { page, pageSize },
				})
			).data,
	);
	// The dummy API persists nothing: keep returned results over the (unchanged) fixtures.
	const [overlay, setOverlay] = useState<Record<string, ApiKey>>({});
	const [dialog, setDialog] = useState<{ key: ApiKey; open: boolean }>();
	const keys = (list.data?.items ?? []).map((key) => overlay[key.id] ?? key);
	return (
		<>
			<ApiKeyList
				keys={keys}
				{...sectionOf(list)}
				pagination={list.data?.pagination}
				loading={list.isFetching}
				onPage={(next) => {
					void navigate({ search: (previous) => ({ ...previous, page: next }) });
				}}
				onRetry={() => {
					void list.refetch();
				}}
				onRevoke={(key) => setDialog({ key, open: true })}
			/>
			{dialog && (
				<AdjustmentDialog
					{...revokeApiKeyDialog(dialog.key)}
					open={dialog.open}
					onOpenChange={(open) => setDialog((current) => current && { ...current, open })}
					onSubmit={(reason) =>
						adjust("org-api-keys", organizationId, async (t) => {
							const response = await sdk.boV1RevokeApiKey({
								...(await t.writeOptions()),
								path: { organizationId, apiKeyId: dialog.key.id },
								body: { reason },
							});
							setOverlay((current) => ({ ...current, [response.data.result.id]: response.data.result }));
							return response.data;
						})
					}
					renderResult={revokeApiKeyResult}
				/>
			)}
		</>
	);
}
