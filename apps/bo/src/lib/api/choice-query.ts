import { useAutocompleteInfiniteQuery } from "@repo/ui/autocomplete/query";
import type { AutocompleteProps } from "@repo/ui/autocomplete";
import type { BOTransport } from "./client";
import { errorMessage } from "./client";
import { useProtected } from "./session";

export type ChoiceBinding<Item> = Extract<AutocompleteProps<Item>, { mode: "selection"; multiple?: false }>;

/** Bind the shared query adapter to BO's authenticated, cancellable SDK transport. */
export function usePrivateChoices<
	Page extends { pagination: { page: number; pageSize: number; total: number } },
	Item extends { id: string; name: string },
>(
	key: readonly unknown[],
	read: (transport: BOTransport, signal: AbortSignal, page: number, search: string) => Promise<Page>,
	getItems: (page: Page) => readonly Item[],
	enabled = true,
): { props: ChoiceBinding<Item>; setSearch: (search: string) => void } {
	const { staff, transport, epoch } = useProtected();
	const result = useAutocompleteInfiniteQuery<Page, Item, Page, Error, readonly unknown[], number, "selection", false>({
		mode: "selection",
		multiple: false,
		enabled,
		queryOptions: ({ search }) => ({
			queryKey: ["private", staff.id, epoch, ...key, search],
			initialPageParam: 1,
			queryFn: ({ signal, pageParam }) => read(transport, signal, pageParam, search),
			getNextPageParam: (page) =>
				page.pagination.page * page.pagination.pageSize < page.pagination.total ? page.pagination.page + 1 : undefined,
			retry: false,
		}),
		getItems,
		getItemValue: (item) => item.id,
		formatError: errorMessage,
	});
	// Let Autocomplete synchronize its displayed text with externally hydrated selections.
	// The query adapter still receives edits through onInputValueChange.
	const { inputValue: _inputValue, ...binding } = result.autocompleteProps;
	const props: ChoiceBinding<Item> = {
		...binding,
		labelStyle: "floating",
		getItemLabel: (item) => item.name,
		getItemValue: (item) => item.id,
		clearable: true,
		clearLabel: "Kosongkan pilihan",
		retryLabel: "Muat ulang",
		loadingMessage: "Memuat pilihan…",
		emptyMessage: "Tidak ada pilihan yang cocok",
		pagination: {
			...result.pagination,
			loadMoreLabel: "Muat pilihan lainnya",
			retryLabel: "Coba lagi",
			loadingMessage: "Memuat pilihan lainnya…",
		},
	};
	return { props, setSearch: result.setSearch };
}
