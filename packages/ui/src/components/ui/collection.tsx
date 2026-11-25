"use client";

import { useRender } from "@base-ui/react/use-render";
import { cn } from "cn";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { ReactElement, ReactNode, RefObject } from "react";

/**
 * Where a column's value goes in a stacked row or a card (a table keeps every column in order):
 * - `title`: the record's heading. Defaults to the first column.
 * - `media`: an avatar or thumbnail beside the title. Not a column of its own.
 * - `trailing`: the value people compare (money, counts), at the end. Defaults for `align: "end"`.
 * - `meta`: secondary values (status, dates, method). The default for the rest.
 * - `actions`: row controls.
 * - `hidden`: wide table only; leave out values the detail page shows anyway.
 */
export type CollectionSlot = "title" | "media" | "trailing" | "meta" | "actions" | "hidden";

export type CollectionColumn<Item> = {
	/** Stable key for the column; also the sort key. */
	id: string;
	header: string;
	cell: (item: Item) => ReactNode;
	/** `end` right-aligns with tabular figures; use it for money and counts. */
	align?: "start" | "end";
	slot?: CollectionSlot;
	/**
	 * Show the header beside the value in stacked rows and cards. Off by default: badges, money and methods explain
	 * themselves. Turn it on where a value is ambiguous without its header: several dates, plain numbers, codes.
	 */
	inlineLabel?: boolean;
	/** The table header sorts by this column when the collection has `onSortChange`. */
	sortable?: boolean;
};

export type CollectionSort = { id: string; direction: "asc" | "desc" };

export type CollectionLayout = "table" | "cards";

export type CollectionProps<Item> = {
	/** Accessible name: the table caption or the list's label. */
	label: string;
	items: readonly Item[];
	columns: readonly CollectionColumn<Item>[];
	getKey: (item: Item) => string;
	/**
	 * The record's page as a link element without children, such as a router `Link`. It wraps the title and covers the
	 * whole row or card. Return nothing for records without a page: they do not look clickable.
	 */
	getLink?: (item: Item) => ReactElement | undefined;
	/**
	 * Both are a list of records; they differ in how each record is laid out. `table` (default): records with
	 * attributes, a table when the collection's own width fits and a one-column list of stacked rows when narrow.
	 * `cards`: records with a face (an avatar or image), a multi-column list of cards.
	 */
	layout?: CollectionLayout;
	sort?: CollectionSort;
	onSortChange?: (sort: CollectionSort) => void;
	labels?: { sortAscending: string; sortDescending: string };
	className?: string;
};

type Column<Item> = CollectionColumn<Item> & { align: "start" | "end"; slot: CollectionSlot };

// Matches Tailwind's `@2xl` container (42rem): the same threshold before and after measuring.
const wideRem = 42;

/** `undefined` until measured; then whether the element is at least the wide container width. */
const useWide = (ref: RefObject<HTMLElement | null>, enabled: boolean) => {
	const [wide, setWide] = useState<boolean>();
	useEffect(() => {
		const element = ref.current;
		if (!enabled || !element || typeof ResizeObserver === "undefined") return;
		const threshold = () => wideRem * parseFloat(getComputedStyle(document.documentElement).fontSize || "16");
		const observer = new ResizeObserver(([entry]) => {
			if (entry) setWide(entry.contentRect.width >= threshold());
		});
		observer.observe(element);
		return () => observer.disconnect();
	}, [ref, enabled]);
	return wide;
};

// The title link stretches over its row or card (the nearest positioned ancestor) through ::after: native link
// behaviour for the whole surface (middle/modifier click, focus), no click handler. Focus shows on the surface.
const stretchedLink =
	"rounded-sm outline-none after:absolute after:inset-0 focus-visible:after:ring-2 focus-visible:after:ring-ring/60 focus-visible:after:ring-inset";
// Controls outside the title sit above the stretched link and stay independently usable.
const raised =
	"[&_:is(a,button,input,select,textarea,label,[role=button],[role=checkbox],[role=switch],[tabindex])]:relative [&_:is(a,button,input,select,textarea,label,[role=button],[role=checkbox],[role=switch],[tabindex])]:z-10";
// Surfaces only look interactive when they open something.
const interactive = "transition-colors hover:bg-muted/50 motion-reduce:transition-none";

/** A value that is not there: a table shows a dash, stacked rows and cards leave it out. */
const isEmpty = (content: ReactNode) => content === null || content === undefined || content === false || content === "";

const TitleLink = ({ link, children }: { link: ReactElement; children: ReactNode }) =>
	useRender({ render: link, props: { className: stretchedLink, children } });

const Title = ({ link, children }: { link?: ReactElement; children: ReactNode }) =>
	link ? <TitleLink link={link}>{children}</TitleLink> : <>{children}</>;

const Value = ({
	header,
	labelClass,
	className,
	children,
}: {
	header: string;
	labelClass: string;
	className?: string;
	children: ReactNode;
}) => (
	<div className={cn("flex min-w-0 items-baseline gap-1.5", className)}>
		<dt className={labelClass}>{header}</dt>
		<dd className="min-w-0 wrap-anywhere">{children}</dd>
	</div>
);

const TableValue = ({ content }: { content: ReactNode }) =>
	isEmpty(content) ? <span className="text-muted-foreground">—</span> : <>{content}</>;

const SortButton = ({
	column,
	sort,
	onSortChange,
	labels,
}: {
	column: { id: string; header: string; align: "start" | "end" };
	sort?: CollectionSort;
	onSortChange: (sort: CollectionSort) => void;
	labels: NonNullable<CollectionProps<unknown>["labels"]>;
}) => {
	const active = sort?.id === column.id ? sort.direction : undefined;
	const Icon = active === "asc" ? ArrowUp : active === "desc" ? ArrowDown : ArrowUpDown;
	return (
		<button
			type="button"
			onClick={() => onSortChange({ id: column.id, direction: active === "asc" ? "desc" : "asc" })}
			aria-label={`${column.header}, ${active === "asc" ? labels.sortDescending : labels.sortAscending}`}
			className={cn(
				"-mx-1.5 inline-flex h-7 items-center gap-1 rounded-md px-1.5 transition-colors outline-none hover:bg-accent hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50 motion-reduce:transition-none",
				active && "text-foreground",
				column.align === "end" && "flex-row-reverse",
			)}
		>
			{column.header}
			<Icon aria-hidden="true" className={cn("size-3.5", !active && "opacity-50")} />
		</button>
	);
};

/**
 * A record collection from one column definition, shown as a table or cards. A table decides by the collection's own
 * width (a container query, then a measurement that keeps a single layout mounted), so a narrow pane behaves like a
 * phone: there the records stack as a one-column list.
 */
export function Collection<Item>({
	label,
	items,
	columns,
	getKey,
	getLink,
	layout = "table",
	sort,
	onSortChange,
	labels = { sortAscending: "sort ascending", sortDescending: "sort descending" },
	className,
}: CollectionProps<Item>) {
	const ref = useRef<HTMLDivElement>(null);
	const wide = useWide(ref, layout === "table");
	const cols: Column<Item>[] = columns.map((column, index) => {
		const align = column.align ?? "start";
		return { ...column, align, slot: column.slot ?? (index === 0 ? "title" : align === "end" ? "trailing" : "meta") };
	});
	const title = cols.find((column) => column.slot === "title") ?? cols[0];
	const media = cols.find((column) => column.slot === "media");
	const of = (slot: CollectionSlot) => cols.filter((column) => column.slot === slot && column !== title);
	const meta = of("meta");
	const trailing = of("trailing");
	const actions = of("actions");
	const tableColumns = cols.filter((column) => column.slot !== "media");
	// Before measuring, both auto layouts render and CSS shows one; afterwards only the measured one mounts.
	const showTable = layout === "table" && wide !== false;
	const showRows = layout === "table" && wide !== true;
	if (!title) return null;
	const labelClass = (column: Column<Item>) => (column.inlineLabel ? "shrink-0 text-muted-foreground" : "sr-only");
	// Stacked rows and cards leave out a missing value; a table shows a dash (`TableValue`).
	const values = (item: Item, list: readonly Column<Item>[], className?: (column: Column<Item>) => string | undefined) =>
		list.map((column) => {
			const content = column.cell(item);
			if (isEmpty(content)) return null;
			return (
				<Value key={column.id} header={column.header} labelClass={labelClass(column)} className={className?.(column)}>
					{content}
				</Value>
			);
		});
	const heading = (item: Item, className?: string) => (
		<div data-collection-title className={cn("flex min-w-0 items-center gap-3", className)}>
			{media && <div className="shrink-0">{media.cell(item)}</div>}
			<div className="min-w-0 flex-1 text-sm wrap-anywhere [&>a]:font-medium">
				<Title link={getLink?.(item)}>{title.cell(item)}</Title>
			</div>
		</div>
	);

	return (
		<div ref={ref} data-slot="collection" data-layout={layout} className={cn("@container min-w-0", className)}>
			{showTable && (
				<div className={cn("min-w-0 overflow-x-auto rounded-xl border border-border", wide === undefined && "hidden @2xl:block")}>
					<table className="w-full border-collapse text-left text-sm">
						<caption className="sr-only">{label}</caption>
						<thead>
							<tr className="border-b border-border bg-muted/40">
								{tableColumns.map((column) => (
									<th
										key={column.id}
										scope="col"
										aria-sort={
											sort?.id === column.id
												? sort.direction === "asc"
													? "ascending"
													: "descending"
												: column.sortable && onSortChange
													? "none"
													: undefined
										}
										className={cn(
											"h-9 px-3 text-xs font-medium whitespace-nowrap text-muted-foreground first:ps-4 last:pe-4",
											column.align === "end" && "text-right",
										)}
									>
										{column.sortable && onSortChange ? (
											<SortButton column={column} sort={sort} onSortChange={onSortChange} labels={labels} />
										) : (
											column.header
										)}
									</th>
								))}
							</tr>
						</thead>
						<tbody>
							{items.map((item) => {
								const link = getLink?.(item);
								return (
									<tr key={getKey(item)} className={cn("relative border-b border-border last:border-0", link && interactive)}>
										{tableColumns.map((column) => (
											<td
												key={column.id}
												className={cn(
													"px-3 py-2.5 align-middle first:ps-4 last:pe-4",
													column.align === "end" && "text-right whitespace-nowrap tabular-nums",
													column === title ? "max-w-96 min-w-48 wrap-anywhere" : raised,
												)}
											>
												{column === title ? heading(item) : <TableValue content={column.cell(item)} />}
											</td>
										))}
									</tr>
								);
							})}
						</tbody>
					</table>
				</div>
			)}
			{showRows && (
				<ul
					aria-label={label}
					className={cn("divide-y divide-border overflow-hidden rounded-xl border border-border", wide === undefined && "@2xl:hidden")}
				>
					{items.map((item) => {
						const link = getLink?.(item);
						return (
							<li
								key={getKey(item)}
								// Title | trailing, then meta, then actions.
								className={cn("relative grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1.5 px-4 py-3", link && interactive)}
							>
								{heading(item, "col-start-1 row-start-1")}
								{meta.length > 0 && (
									<dl className={cn("col-span-2 row-start-2 flex min-w-0 flex-wrap items-center gap-x-4 gap-y-1 text-xs", raised)}>
										{values(item, meta)}
									</dl>
								)}
								{trailing.length > 0 && (
									<dl
										className={cn("col-start-2 row-start-1 flex flex-col items-end gap-0.5 text-sm whitespace-nowrap tabular-nums", raised)}
									>
										{values(item, trailing)}
									</dl>
								)}
								{actions.length > 0 && (
									<div className={cn("col-span-2 row-start-3 flex flex-wrap items-center gap-2 pt-1", raised)}>
										{actions.map((column) => (
											<div key={column.id}>{column.cell(item)}</div>
										))}
									</div>
								)}
							</li>
						);
					})}
				</ul>
			)}
			{layout === "cards" && (
				<ul aria-label={label} className="grid grid-cols-[repeat(auto-fill,minmax(min(100%,17rem),1fr))] gap-3">
					{items.map((item) => {
						const link = getLink?.(item);
						return (
							<li
								key={getKey(item)}
								className={cn("relative flex flex-col gap-3 rounded-xl border border-border bg-background p-4", link && interactive)}
							>
								<div className="flex min-w-0 items-start justify-between gap-3">
									{heading(item, "flex-1")}
									{trailing.length > 0 && (
										<dl className={cn("flex shrink-0 flex-col items-end gap-0.5 text-sm whitespace-nowrap tabular-nums", raised)}>
											{values(item, trailing)}
										</dl>
									)}
								</div>
								{meta.length > 0 && <dl className={cn("flex min-w-0 flex-col gap-1.5 text-xs", raised)}>{values(item, meta)}</dl>}
								{actions.length > 0 && (
									<div className={cn("mt-auto flex flex-wrap items-center gap-2", raised)}>
										{actions.map((column) => (
											<div key={column.id}>{column.cell(item)}</div>
										))}
									</div>
								)}
							</li>
						);
					})}
				</ul>
			)}
		</div>
	);
}
