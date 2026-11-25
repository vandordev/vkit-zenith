import { useState } from "react";
import { usePrivateQuery } from "@bo/lib/api/session";
import { boV1ListUserChoices } from "@bo/lib/api/generated/sdk.gen";
import { errorMessage } from "@bo/lib/api/client";
import type { V1UserChoiceListItem } from "@bo/lib/api/generated/types.gen";
import type { CustomerOwnerChoices } from "../-components/organization-create";
export function useCustomerChoices(enabled: boolean, organizationId?: string): CustomerOwnerChoices {
	const [q, setQ] = useState("");
	const [page, setPage] = useState(1);
	const [selected, setSelected] = useState<V1UserChoiceListItem | null>(null);
	const choices = usePrivateQuery(
		["customer-choices", organizationId, q, page, selected?.id],
		async (t, signal) =>
			(
				await boV1ListUserChoices({
					client: t.client,
					signal,
					throwOnError: true,
					query: { q, page, pageSize: 25, selectedId: selected?.id, organizationId },
				})
			).data,
		enabled,
	);
	return {
		data: choices.data,
		value: choices.data ? choices.data.selected : selected,
		q,
		page,
		loading: choices.isFetching,
		error: choices.error ? errorMessage(choices.error) : undefined,
		onSelected: setSelected,
		onSearch: (next) => {
			setQ(next);
			setPage(1);
		},
		onPage: setPage,
		onRetry: () => {
			void choices.refetch();
		},
	};
}
