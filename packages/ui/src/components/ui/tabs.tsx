"use client";

import { Tabs as TabsPrimitive } from "@base-ui/react/tabs";
import { useRender } from "@base-ui/react/use-render";
import { cn } from "cn";
import { motion as brandMotion } from "@repo/brand/motion";
import { motion, useReducedMotion } from "motion/react";
import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import type { ComponentProps, ReactElement, ReactNode, RefObject } from "react";

type TabsVariant = "underline" | "pill" | "segmented";

interface TabsPanelItem<Value extends string = string> {
	value: Value;
	label: ReactNode;
	content: ReactNode;
	disabled?: boolean;
	link?: never;
	active?: never;
}

interface TabsLinkItem {
	link: ReactElement;
	active?: boolean;
	/** Optional stable identity for dynamic or reordered link collections. */
	key?: string;
	value?: never;
	label?: never;
	content?: never;
	disabled?: never;
}

type TabsCommonProps = Omit<ComponentProps<"div">, "children" | "defaultValue" | "ref"> & {
	variant?: TabsVariant;
	/** `sm` is the compact trigger for dense bars, such as an app header. */
	size?: "default" | "sm";
	orientation?: "horizontal" | "vertical";
	animated?: boolean;
	listClassName?: string;
};

type TabsPanelProps<Value extends string = string> = TabsCommonProps & {
	items: readonly TabsPanelItem<Value>[];
	value?: Value | null;
	defaultValue?: Value | null;
	onValueChange?: (value: Value | null, details: TabsPrimitive.Root.ChangeEventDetails) => void;
	activationMode?: "manual" | "automatic";
	keepMounted?: boolean;
	contentClassName?: string;
	ref?: ComponentProps<"div">["ref"];
};

type TabsNavigationProps = TabsCommonProps & {
	items: readonly TabsLinkItem[];
	ref?: ComponentProps<"nav">["ref"];
	value?: never;
	defaultValue?: never;
	onValueChange?: never;
	activationMode?: never;
	keepMounted?: never;
	contentClassName?: never;
};

type TabsProps<Value extends string = string> = TabsPanelProps<Value> | TabsNavigationProps;

const transition = { duration: brandMotion.base, ease: brandMotion.easeOut };

// Overflowing tabs scroll sideways. A thin bar keeps the height jump small when it appears in a narrow
// or resizing container (a classic bar adds ~15px and arrows); wheel, touch and keyboard still scroll.
const scrollerClassName = "max-w-full overflow-x-auto overflow-y-hidden p-1 [scrollbar-width:thin]";

const listClasses = (variant: TabsVariant, vertical: boolean) =>
	cn(
		"relative isolate flex w-max items-center",
		vertical ? "flex-col items-stretch" : "min-w-0",
		variant === "underline" && (vertical ? "gap-1 border-s border-border" : "gap-1 border-b border-border"),
		variant === "pill" && "gap-1",
		variant === "segmented" && "gap-1 rounded-lg bg-muted p-1",
	);

const triggerClasses = (variant: TabsVariant, size: "default" | "sm" = "default") =>
	cn(
		"relative inline-flex shrink-0 items-center justify-center gap-2 font-medium whitespace-nowrap text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-[3px] focus-visible:ring-ring/50 data-active:text-foreground data-disabled:pointer-events-none data-disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
		size === "sm" ? "min-h-7 px-2.5 py-1 text-xs [&_svg]:size-3.5" : "min-h-9 px-3 py-2 text-sm",
		// Segmented triggers sit concentric inside the list: rounded-lg minus its 4px padding is rounded-sm.
		variant === "underline" || variant === "segmented" ? "rounded-sm" : "rounded-md",
	);

const Indicator = ({ variant, vertical, animated }: { variant: TabsVariant; vertical: boolean; animated: boolean }) => {
	const reduceMotion = useReducedMotion();
	const enabled = animated && !reduceMotion;
	// A single local canvas never projects between trigger parents. Moving a
	// centered preview or resizing its panel cannot add a viewport-space detour.
	const ref = useRef<SVGSVGElement>(null);
	const [box, setBox] = useState<{
		x: number;
		y: number;
		width: number;
		height: number;
		/** Glide only when the active tab changed; a resize (window, animating sidebar, font swap) snaps,
		 * otherwise the indicator trails its tab and settles after the container stops moving. */
		glide: boolean;
	} | null>(null);
	const lastActive = useRef<Element | null>(null);
	useLayoutEffect(() => {
		const list = ref.current?.parentElement;
		if (!list) {
			return;
		}
		const win = list.ownerDocument.defaultView;
		if (!win) {
			return;
		}
		const measure = () => {
			const active = list.querySelector<HTMLElement>(
				// Links use data-active (the item's own `active`), not aria-current: a router Link also sets
				// aria-current on every route it matches, so a parent tab ("Semua") would win document order.
				'[data-slot="tabs-trigger"][aria-selected="true"], [data-slot="tabs-link"][data-active]',
			);
			if (!active) {
				lastActive.current = null;
				setBox(null);
				return;
			}
			const bounds = list.getBoundingClientRect();
			const tab = active.getBoundingClientRect();
			const scaleX = list.offsetWidth && bounds.width ? bounds.width / list.offsetWidth : 1;
			const scaleY = list.offsetHeight && bounds.height ? bounds.height / list.offsetHeight : 1;
			let width = tab.width / scaleX;
			let height = tab.height / scaleY;
			let x = (tab.left - bounds.left) / scaleX - list.clientLeft + list.scrollLeft;
			let y = (tab.top - bounds.top) / scaleY - list.clientTop + list.scrollTop;
			const underline = variant === "underline";
			if (underline) {
				if (vertical) {
					x = win.getComputedStyle(list).direction === "rtl" ? (list.clientWidth || bounds.width / scaleX) - 2 : 0;
					y += 8;
					width = 2;
					height = Math.max(0, height - 16);
				} else {
					y = (list.clientHeight || bounds.height / scaleY) - 2;
					height = 2;
				}
			}
			const glide = lastActive.current !== null && lastActive.current !== active;
			lastActive.current = active;
			const next = {
				glide,
				height,
				width,
				x,
				y,
			};
			setBox((previous) =>
				previous && previous.x === next.x && previous.y === next.y && previous.width === next.width && previous.height === next.height
					? previous
					: next,
			);
		};
		const resize = new win.ResizeObserver(measure);
		const observeSize = () => {
			resize.disconnect();
			resize.observe(list);
			for (const tab of list.querySelectorAll('[data-slot="tabs-trigger"], [data-slot="tabs-link"]')) {
				resize.observe(tab);
			}
			measure();
		};
		const mutations = new win.MutationObserver(observeSize);
		mutations.observe(list, {
			attributeFilter: ["aria-selected", "aria-current", "data-active", "dir"],
			attributes: true,
			characterData: true,
			childList: true,
			subtree: true,
		});
		observeSize();
		win.addEventListener("resize", measure);
		return () => {
			mutations.disconnect();
			resize.disconnect();
			win.removeEventListener("resize", measure);
		};
	}, [variant, vertical]);
	return (
		<svg
			ref={ref}
			aria-hidden="true"
			focusable="false"
			data-slot="tabs-indicator"
			className="pointer-events-none absolute inset-0 h-full w-full overflow-hidden"
		>
			{box && box.width > 0 && box.height > 0 && (
				<motion.rect
					initial={false}
					animate={{
						attrX: box.x,
						attrY: box.y,
						height: box.height,
						width: box.width,
					}}
					transition={enabled && box.glide ? transition : { duration: 0 }}
					// Segmented ≈ --radius-sm (12.8px rounded-lg minus the 4px list padding), concentric with the list.
					rx={variant === "underline" ? 0 : variant === "segmented" ? 9 : 6}
					className={cn(
						variant === "underline" && "fill-foreground",
						variant === "pill" && "fill-accent",
						variant === "segmented" && "fill-background stroke-border",
					)}
					vectorEffect="non-scaling-stroke"
				/>
			)}
		</svg>
	);
};

const TabButton = ({ nativeProps }: { nativeProps: ComponentProps<"button"> }) =>
	useRender({
		defaultTagName: "button",
		props: {
			...nativeProps,
			children: <span className="relative z-10 inline-flex items-center gap-2">{nativeProps.children}</span>,
		},
		ref: nativeProps.ref,
	});

const NavigationItem = ({ item, variant, size }: { item: TabsLinkItem; variant: TabsVariant; size?: "default" | "sm" }) => {
	const link = useRender({
		defaultTagName: "a",
		props: {
			"aria-current": item.active ? "page" : undefined,
			className: cn(triggerClasses(variant, size), "z-10 w-full"),
			"data-active": item.active ? "" : undefined,
			"data-slot": "tabs-link",
		},
		render: item.link,
	});
	return <div className="relative shrink-0">{link}</div>;
};

const PanelBody = ({
	children,
	hidden,
	animated,
	hydrated,
}: {
	children: ReactNode;
	hidden: boolean;
	animated: boolean;
	hydrated: RefObject<boolean>;
}) => {
	const reduceMotion = useReducedMotion();
	const enabled = animated && !reduceMotion;
	return (
		<motion.div
			initial={enabled && hydrated.current ? { opacity: 0 } : false}
			animate={{ opacity: hidden ? 0 : 1 }}
			transition={enabled ? { duration: brandMotion.fast, ease: transition.ease } : { duration: 0 }}
		>
			{children}
		</motion.div>
	);
};

const isNavigation = <Value extends string>(props: TabsProps<Value>): props is TabsNavigationProps => props.items[0]?.link !== undefined;

const Tabs = <Value extends string = string>(props: TabsProps<Value>) => {
	const id = useId();
	const hydrated = useRef(false);
	useEffect(() => {
		hydrated.current = true;
	}, []);

	if (props.items.length === 0) {
		return null;
	}

	if (isNavigation(props)) {
		const {
			items,
			variant = "underline",
			size,
			orientation = "horizontal",
			animated = true,
			className,
			listClassName,
			...nativeProps
		} = props;
		return (
			<nav {...nativeProps} data-slot="tabs" data-variant={variant} data-orientation={orientation} className={cn("max-w-full", className)}>
				<div className={scrollerClassName} style={{ overflowY: "hidden" }}>
					<div data-slot="tabs-list" className={cn(listClasses(variant, orientation === "vertical"), listClassName)}>
						{items.map((item, index) => (
							<NavigationItem key={item.key ?? item.link.key ?? index} item={item} variant={variant} size={size} />
						))}
						<Indicator variant={variant} vertical={orientation === "vertical"} animated={animated} />
					</div>
				</div>
			</nav>
		);
	}

	const {
		items,
		variant = "underline",
		size,
		orientation = "horizontal",
		animated = true,
		activationMode = "manual",
		keepMounted = false,
		className,
		listClassName,
		contentClassName,
		defaultValue,
		"aria-label": ariaLabel,
		"aria-labelledby": ariaLabelledBy,
		...rootProps
	} = props as TabsPanelProps<Value>;
	return (
		<TabsPrimitive.Root
			{...rootProps}
			defaultValue={defaultValue === undefined ? (items.find((item) => !item.disabled)?.value ?? null) : defaultValue}
			orientation={orientation}
			data-slot="tabs"
			data-variant={variant}
			className={cn("flex min-w-0 gap-4", orientation === "vertical" ? "flex-row" : "flex-col", className)}
		>
			<div className={cn(scrollerClassName, "shrink-0")} style={{ overflowY: "hidden" }}>
				<TabsPrimitive.List
					aria-label={ariaLabel}
					aria-labelledby={ariaLabelledBy}
					activateOnFocus={activationMode === "automatic"}
					data-slot="tabs-list"
					className={cn(listClasses(variant, orientation === "vertical"), listClassName)}
				>
					{items.map((item) => (
						<TabsPrimitive.Tab
							key={item.value}
							value={item.value}
							id={`${id}-tab-${encodeURIComponent(item.value)}`}
							disabled={item.disabled}
							data-slot="tabs-trigger"
							className={triggerClasses(variant, size)}
							render={(nativeProps, state) => (
								<TabButton
									nativeProps={{
										...nativeProps,
										"aria-controls": state.active || keepMounted ? `${id}-panel-${encodeURIComponent(item.value)}` : undefined,
									}}
								/>
							)}
						>
							{item.label}
						</TabsPrimitive.Tab>
					))}
					<Indicator variant={variant} vertical={orientation === "vertical"} animated={animated} />
				</TabsPrimitive.List>
			</div>
			{items.map((item) => (
				<TabsPrimitive.Panel
					key={item.value}
					value={item.value}
					id={`${id}-panel-${encodeURIComponent(item.value)}`}
					aria-labelledby={`${id}-tab-${encodeURIComponent(item.value)}`}
					keepMounted={keepMounted}
					data-slot="tabs-content"
					className={cn("min-w-0 flex-1 text-sm outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50", contentClassName)}
					render={(nativeProps, state) => (
						<div {...nativeProps}>
							<PanelBody hidden={state.hidden} animated={animated} hydrated={hydrated}>
								{nativeProps.children}
							</PanelBody>
						</div>
					)}
				>
					{item.content}
				</TabsPrimitive.Panel>
			))}
		</TabsPrimitive.Root>
	);
};

export { Tabs };
export type { TabsLinkItem, TabsNavigationProps, TabsPanelItem, TabsPanelProps, TabsProps, TabsVariant };
