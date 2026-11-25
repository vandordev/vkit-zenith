import type { FilterBarProps } from "@repo/ui/filter-bar";

/** Indonesian copy for every BO FilterBar. */
export const filterLabels: NonNullable<FilterBarProps["labels"]> = {
	filter: "Filter",
	filters: "Filter",
	clear: "Kosongkan filter",
	sort: "Urutkan",
	all: "Semua",
	remove: (field) => `Hapus filter ${field.toLowerCase()}`,
	change: (field) => `Ubah filter ${field.toLowerCase()}`,
};
