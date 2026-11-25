"use client";

// Composed from Base UI's Autocomplete "Command palette" example: an inline, always-open autocomplete in a dialog.
// https://base-ui.com/react/components/autocomplete
import { Autocomplete } from "@base-ui/react/autocomplete";
import { Dialog } from "@base-ui/react/dialog";
import { cn } from "cn";
import { SearchIcon } from "lucide-react";
import { useId } from "react";
import type { ReactNode } from "react";

export interface CommandPaletteItem {
	id: string;
	label: string;
	/** Items render under group headings in the order the groups first appear. */
	group: string;
	icon?: ReactNode;
	/** Short trailing text, such as a shortcut or a destination kind. */
	hint?: ReactNode;
}

export interface CommandPaletteProps<Item extends CommandPaletteItem> {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	items: readonly Item[];
	/** Called on click or Enter; the palette closes itself first. */
	onSelect: (item: Item) => void;
	labels?: {
		title: string;
		placeholder: string;
		empty: string;
		close: string;
		activate: string;
		instructions: string;
	};
}

const groupItems = <Item extends CommandPaletteItem>(items: readonly Item[]) => {
	const groups = new Map<string, Item[]>();
	for (const item of items) groups.set(item.group, [...(groups.get(item.group) ?? []), item]);
	return [...groups].map(([value, entries]) => ({ value, items: entries }));
};

export const CommandPalette = <Item extends CommandPaletteItem>({
	open,
	onOpenChange,
	items,
	onSelect,
	labels = {
		title: "Command palette",
		placeholder: "Search pages and commands…",
		empty: "No results.",
		close: "Close command palette",
		activate: "Open",
		instructions: "Use the arrow keys to move and Enter to open the highlighted item.",
	},
}: CommandPaletteProps<Item>) => {
	const instructionsId = useId();
	return (
		<Dialog.Root open={open} onOpenChange={onOpenChange}>
			<Dialog.Portal>
				<Dialog.Backdrop
					data-slot="command-palette-overlay"
					className="fixed inset-0 z-50 bg-black/40 transition-opacity duration-(--motion-base) ease-(--ease-out) data-ending-style:opacity-0 data-starting-style:opacity-0 motion-reduce:transition-none supports-[-webkit-touch-callout:none]:absolute"
				/>
				<Dialog.Viewport className="fixed inset-0 z-50 flex items-start justify-center px-2 pt-[min(12dvh,6rem)] pb-2">
					<Dialog.Popup
						data-slot="command-palette"
						aria-label={labels.title}
						className="flex max-h-[min(32rem,calc(100dvh-5rem))] w-full max-w-xl flex-col overflow-hidden rounded-xl border border-border bg-popover text-popover-foreground shadow-xl outline-none transition-[opacity,scale,translate] duration-(--motion-base) ease-(--ease-out) data-ending-style:-translate-y-2 data-ending-style:scale-[0.98] data-ending-style:opacity-0 data-starting-style:-translate-y-2 data-starting-style:scale-[0.98] data-starting-style:opacity-0 motion-reduce:transition-none"
					>
						<Autocomplete.Root
							open
							inline
							items={groupItems(items)}
							// The group joins the match, so typing a section name ("Settings") lists its pages.
							itemToStringValue={(item: Item) => `${item.label} ${item.group}`}
							autoHighlight="always"
							keepHighlight
						>
							<Autocomplete.InputGroup className="flex shrink-0 items-center gap-2 border-b border-border px-3">
								<SearchIcon aria-hidden="true" className="size-4 shrink-0 text-muted-foreground" />
								<Autocomplete.Input
									data-slot="command-palette-input"
									aria-label={labels.title}
									aria-describedby={instructionsId}
									placeholder={labels.placeholder}
									className="h-12 w-full min-w-0 border-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground any-pointer-coarse:text-base"
								/>
							</Autocomplete.InputGroup>
							<Dialog.Close className="sr-only">{labels.close}</Dialog.Close>
							<span id={instructionsId} className="sr-only">
								{labels.instructions}
							</span>
							<div className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-1.5 [scroll-padding-block:0.375rem]">
								<Autocomplete.Empty>
									<p className="px-3 py-8 text-center text-sm text-muted-foreground">{labels.empty}</p>
								</Autocomplete.Empty>
								<Autocomplete.List>
									{(group: { value: string; items: Item[] }) => (
										<Autocomplete.Group key={group.value} items={group.items} className="not-last:mb-1.5">
											<Autocomplete.GroupLabel className="px-2 pt-1.5 pb-1 text-xs font-medium text-muted-foreground select-none">
												{group.value}
											</Autocomplete.GroupLabel>
											<Autocomplete.Collection>
												{(item: Item) => (
													<Autocomplete.Item
														key={item.id}
														value={item}
														onClick={() => {
															onOpenChange(false);
															onSelect(item);
														}}
														className={cn(
															"flex h-9 cursor-default items-center gap-2.5 rounded-md px-2 text-sm outline-none select-none [scroll-margin-block:0.375rem]",
															"data-highlighted:bg-accent data-highlighted:text-accent-foreground",
														)}
													>
														{item.icon && (
															<span
																aria-hidden="true"
																className="flex size-4 shrink-0 items-center justify-center text-muted-foreground [&_svg]:size-4"
															>
																{item.icon}
															</span>
														)}
														<span className="min-w-0 flex-1 truncate">{item.label}</span>
														{item.hint && <span className="shrink-0 text-xs text-muted-foreground">{item.hint}</span>}
													</Autocomplete.Item>
												)}
											</Autocomplete.Collection>
										</Autocomplete.Group>
									)}
								</Autocomplete.List>
							</div>
							<div className="flex shrink-0 items-center justify-end gap-1.5 border-t border-border px-3 py-2 text-xs text-muted-foreground">
								<span>{labels.activate}</span>
								<kbd className="inline-flex h-5 min-w-5 items-center justify-center rounded border border-border bg-muted px-1 font-sans text-[0.625rem]">
									Enter
								</kbd>
							</div>
						</Autocomplete.Root>
					</Dialog.Popup>
				</Dialog.Viewport>
			</Dialog.Portal>
		</Dialog.Root>
	);
};
