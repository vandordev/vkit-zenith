"use client";

import { cn } from "cn";
import { ArrowDownUp, ListFilter, Plus, X } from "lucide-react";
import type { ReactNode } from "react";

import { Badge } from "./badge";
import { Button } from "./button";
import { Drawer, DrawerBody } from "./drawer";
import { Dropdown, type DropdownEntry } from "./dropdown";
import { InputSearch } from "./input-search";
import { Select } from "./select";

export type FilterOption = { value: string; label: string };

/** One single-choice filter. `""` means no filter (all). */
export type FilterField = {
	id: string;
	label: string;
	options: readonly FilterOption[];
	value: string;
	onChange: (value: string) => void;
};

export type FilterBarProps = {
	/** The consumer owns debouncing (it holds the draft) and what the text matches. */
	search?: { value: string; onChange: (value: string) => void; placeholder: string; label?: string };
	filters?: readonly FilterField[];
	/** Only for orders the data source supports: sorting a single page of results would mislead. */
	sort?: { value: string; options: readonly FilterOption[]; onChange: (value: string) => void };
	/** Shown when a filter or search text is active. */
	onClear?: () => void;
	/** Extra controls at the end of the bar, such as a secondary action. */
	end?: ReactNode;
	labels?: {
		filter: string;
		filters: string;
		clear: string;
		sort: string;
		all: string;
		remove: (field: string) => string;
		change: (field: string) => string;
	};
	className?: string;
};

const defaultLabels: NonNullable<FilterBarProps["labels"]> = {
	filter: "Filter",
	filters: "Filters",
	clear: "Clear filters",
	sort: "Sort",
	all: "All",
	remove: (field) => `Remove ${field} filter`,
	change: (field) => `Change ${field} filter`,
};

const radio = (field: FilterField, all: string): DropdownEntry => ({
	id: field.id,
	type: "radio-group",
	value: field.value,
	onValueChange: field.onChange,
	items: [
		{ id: "", value: "", label: all },
		...field.options.map((option) => ({ id: option.value, value: option.value, label: option.label })),
	],
});

/**
 * The toolbar above a collection: search, a "Filter" menu that adds single-choice filters, the active filters as
 * removable chips, and an optional sort. When the bar is narrow, filters and sort move into a drawer.
 */
export function FilterBar({ search, filters = [], sort, onClear, end, labels = defaultLabels, className }: FilterBarProps) {
	const active = filters.filter((field) => field.value !== "");
	const optionLabel = (field: FilterField) => field.options.find((option) => option.value === field.value)?.label ?? field.value;
	const sortLabel = sort?.options.find((option) => option.value === sort.value)?.label;
	const showClear = Boolean(onClear) && (active.length > 0 || Boolean(search?.value));
	return (
		<div data-slot="filter-bar" className={cn("@container flex min-w-0 flex-col gap-2", className)}>
			<div className="flex min-w-0 items-center gap-2">
				{search && (
					<div className="min-w-0 flex-1 @xl:max-w-sm">
						<InputSearch
							aria-label={search.label ?? search.placeholder}
							placeholder={search.placeholder}
							value={search.value}
							onChange={(event) => search.onChange(event.target.value)}
							clearable
							maxLength={100}
							className="h-9"
						/>
					</div>
				)}
				{filters.length > 0 && (
					<Dropdown
						trigger={
							<Button type="button" variant="outline" size="sm" className="hidden h-9 shrink-0 @xl:inline-flex">
								<Plus aria-hidden="true" />
								{labels.filter}
							</Button>
						}
						items={filters.map((field) => ({ id: field.id, type: "submenu", label: field.label, items: [radio(field, labels.all)] }))}
						contentProps={{ align: "start" }}
					/>
				)}
				<div className="ms-auto flex shrink-0 items-center gap-2">
					{sort && (
						<Dropdown
							trigger={
								<Button type="button" variant="ghost" size="sm" className="hidden h-9 @xl:inline-flex">
									<ArrowDownUp aria-hidden="true" />
									<span className="sr-only">{labels.sort}: </span>
									{sortLabel ?? labels.sort}
								</Button>
							}
							items={[
								{
									id: "sort",
									type: "radio-group",
									label: labels.sort,
									value: sort.value,
									onValueChange: sort.onChange,
									items: sort.options.map((option) => ({ id: option.value, ...option })),
								},
							]}
							contentProps={{ align: "end" }}
						/>
					)}
					{(filters.length > 0 || sort) && (
						<Drawer
							title={labels.filters}
							swipeDirection="down"
							trigger={
								<Button type="button" variant="outline" size="icon" className="relative @xl:hidden" aria-label={labels.filters}>
									<ListFilter aria-hidden="true" />
									{active.length > 0 && (
										<Badge size="xs" className="absolute -top-1.5 -right-1.5">
											{active.length}
										</Badge>
									)}
								</Button>
							}
						>
							<DrawerBody className="flex flex-col gap-4">
								{filters.map((field) => (
									<Select
										key={field.id}
										label={field.label}
										labelStyle="static"
										className="w-full"
										value={field.value}
										data={[{ value: "", label: labels.all }, ...field.options]}
										onValueChange={(value) => field.onChange(typeof value === "string" ? value : "")}
									/>
								))}
								{sort && (
									<Select
										label={labels.sort}
										labelStyle="static"
										className="w-full"
										value={sort.value}
										data={[...sort.options]}
										onValueChange={(value) => typeof value === "string" && sort.onChange(value)}
									/>
								)}
							</DrawerBody>
						</Drawer>
					)}
					{end}
				</div>
			</div>
			{(active.length > 0 || showClear) && (
				<div className="flex min-w-0 flex-wrap items-center gap-1.5">
					{active.map((field) => (
						<div
							key={field.id}
							className="inline-flex h-7 min-w-0 items-center overflow-hidden rounded-md border border-border bg-background text-xs"
						>
							<Dropdown
								trigger={
									<button
										type="button"
										aria-label={labels.change(field.label)}
										className="inline-flex h-full min-w-0 items-center gap-1 ps-2 pe-1.5 outline-none hover:bg-accent focus-visible:bg-accent"
									>
										<span className="text-muted-foreground">{field.label}</span>
										<span className="truncate font-medium">{optionLabel(field)}</span>
									</button>
								}
								items={[radio(field, labels.all)]}
								contentProps={{ align: "start" }}
							/>
							<button
								type="button"
								aria-label={labels.remove(field.label)}
								onClick={() => field.onChange("")}
								className="inline-flex h-full items-center border-s border-border px-1.5 text-muted-foreground outline-none hover:bg-accent hover:text-foreground focus-visible:bg-accent"
							>
								<X aria-hidden="true" className="size-3.5" />
							</button>
						</div>
					))}
					{showClear && (
						<Button type="button" variant="ghost" size="xs" onClick={onClear} className="text-muted-foreground">
							{labels.clear}
						</Button>
					)}
				</div>
			)}
		</div>
	);
}
