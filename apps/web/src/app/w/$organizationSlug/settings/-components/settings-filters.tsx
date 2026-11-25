import { FilterBar } from "@repo/ui/filter-bar";
import { filterLabels } from "../../-components/filter-labels";
import { useSearchDraft } from "../../-components/search-draft";

export type SettingsFilterValue = { q: string; status: string };
export type StatusOption = { value: string; label: string };

/**
 * Same behaviour as the catalog filters: status applies at once, the text is debounced into the URL, and `navigationKey` (page, status and
 * mode) drops a pending draft on a history move. The status options belong to the list (the API's enum for that menu).
 */
export function SettingsFilters({
	filters,
	onFilter,
	searchLabel,
	statusOptions,
	navigationKey,
}: {
	filters: SettingsFilterValue;
	onFilter: (filters: SettingsFilterValue) => void;
	searchLabel: string;
	statusOptions: readonly StatusOption[];
	navigationKey: string;
}) {
	const search = useSearchDraft(filters.q, (q) => onFilter({ ...filters, q }), navigationKey);
	return (
		<FilterBar
			search={{ value: search.draft, onChange: search.change, placeholder: `${searchLabel}…`, label: searchLabel }}
			filters={[
				{
					id: "status",
					label: "Status",
					options: statusOptions,
					value: filters.status,
					onChange: (status) => {
						if (status === "" || statusOptions.some((option) => option.value === status)) onFilter({ q: search.settle(), status });
					},
				},
			]}
			onClear={() => onFilter({ q: search.settle(""), status: "" })}
			labels={filterLabels}
		/>
	);
}
