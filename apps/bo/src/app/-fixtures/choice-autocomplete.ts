import { useState } from "react";
import type { ChoiceBinding } from "@bo/lib/api/choice-query";

/** Local presentation fixtures only; never imports a transport or live SDK. */
export function useChoiceFixture<Item extends { id: string; name: string }>(
	allItems: readonly Item[],
	error?: string,
	loading = false,
): ChoiceBinding<Item> {
	const [search, setSearch] = useState("");
	const [page, setPage] = useState(1);
	const [retried, setRetried] = useState(false);
	const filtered = allItems.filter((item) => item.name.toLowerCase().includes(search.toLowerCase()));
	const props: ChoiceBinding<Item> = {
		mode: "selection",
		labelStyle: "floating",
		items: filtered.slice(0, page * 25),
		getItemLabel: (item) => item.name,
		getItemValue: (item) => item.id,
		filter: null,
		onInputValueChange: (next) => {
			setSearch(next);
			setPage(1);
		},
		clearable: true,
		clearLabel: "Kosongkan pilihan",
		loading,
		error: retried ? undefined : error,
		onRetry: () => setRetried(true),
		retryLabel: "Muat ulang",
		loadingMessage: "Memuat pilihan…",
		emptyMessage: "Tidak ada pilihan yang cocok",
		pagination: {
			hasNextPage: page * 25 < filtered.length,
			fetchingNextPage: false,
			onLoadMore: () => setPage((p) => p + 1),
			loadMoreLabel: "Muat pilihan lainnya",
		},
	};
	return props;
}
