"use client";

// Adapted from Universe's AppShell, originally inspired by 7ovr's
// app-shell-3 icon-rail layout (free blocks: MIT-0).
// https://7ovr.com/r/app-shell-3.json
// Application data, routing and resource controls remain consumer-owned.
import { cn } from "cn";
import { motion as brandMotion } from "@repo/brand/motion";
import { MenuIcon, PanelLeftCloseIcon, PanelLeftOpenIcon, SearchIcon } from "lucide-react";
import { AnimatePresence, motion, useIsPresent, useReducedMotion } from "motion/react";
import { isValidElement, useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import type { ReactNode } from "react";

import { Button } from "./button";
import { Drawer, DrawerBody } from "./drawer";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "./tooltip";
import { zenithWordmarkClassName } from "./zenith-mark";

export interface AppShellNavigationState {
	collapsed: boolean;
	closeNavigation: () => void;
}

export interface AppShellNavigationItem {
	id: string;
	label: string;
	icon?: ReactNode;
	href?: string;
	onClick?: () => void;
	active?: boolean;
	disabled?: boolean;
	/** Separates this item from the following navigation. */
	dividerAfter?: boolean;
	/** Replaces the standard item. The caller owns semantics and disabled state. */
	renderControl?: (state: AppShellNavigationState) => ReactNode;
	/** A collapsed rail shows the label as a tooltip on `renderControl` items too; set false for non-interactive rows (headings). */
	tooltip?: boolean;
}

export interface AppShellProps {
	labels?: {
		skipToContent: string;
		sidebar: string;
		openNavigation: string;
		closeNavigation: string;
		expandSidebar: string;
		collapseSidebar: string;
	};
	/** `name` is the accessible name; `label` replaces the visible text beside the icon. */
	brand: { name: string; icon?: ReactNode; label?: ReactNode; href?: string };
	navigation: readonly AppShellNavigationItem[];
	children: ReactNode;
	/** Hide the topbar and move desktop controls into the sidebar. */
	showHeader?: boolean;
	/**
	 * Content sits in a rounded panel on the sidebar surface (Linear-style). The desktop collapse toggle and
	 * search move into the sidebar; the header, if shown, is a slim bar inside the panel. On desktop the panel
	 * scrolls, not the window: route scroll reset must target `[data-scroll-restoration-id="app-shell-main"]`.
	 */
	inset?: boolean;
	header?: ReactNode;
	actions?: ReactNode;
	sidebarHeader?: ReactNode | ((state: AppShellNavigationState) => ReactNode);
	sidebarFooter?: ReactNode | ((state: AppShellNavigationState) => ReactNode);
	/** Enables the search button and Cmd/Ctrl+K listener only when supplied. */
	onSearch?: () => void;
	searchLabel?: string;
	defaultCollapsed?: boolean;
	collapsed?: boolean;
	onCollapsedChange?: (collapsed: boolean) => void;
	navigationLabel?: string;
	/** Change key only for a new panel, not active items or same-level resources. */
	navigationContext?: { key: string; depth: number };
	/**
	 * Drill-down header at the top of the navigation panel: back control, title, meta slot.
	 * It slides with the panel; identity and animation stay owned by `navigationContext`.
	 */
	sidebarContext?: {
		title: string;
		back: { label: string; render: (state: AppShellNavigationState) => ReactNode };
		meta?: ReactNode;
	};
	className?: string;
	/** Main is edge-to-edge; page content owns its padding (normally PageContainer). */
	contentClassName?: string;
}

/**
 * Entity name limited to two balanced lines. Only when it is actually clamped does it become a
 * focusable tooltip trigger, so the full name stays reachable by hover and keyboard without adding a
 * tab stop to short names.
 */
const ContextTitle = ({ title }: { title: string }) => {
	const ref = useRef<HTMLHeadingElement>(null);
	const [clamped, setClamped] = useState(false);
	useLayoutEffect(() => {
		const element = ref.current;
		if (!element) return;
		let active = true;
		const measure = () => {
			if (active) setClamped(element.scrollHeight > element.clientHeight + 1);
		};
		measure();
		const observer = new ResizeObserver(measure);
		observer.observe(element);
		// A late font swap can change the text width without changing the clamped box height, so the
		// observer stays silent: remeasure once fonts settle and after every later font load.
		const fonts = typeof document === "undefined" ? undefined : document.fonts;
		void fonts?.ready.then(measure);
		fonts?.addEventListener("loadingdone", measure);
		return () => {
			active = false;
			observer.disconnect();
			fonts?.removeEventListener("loadingdone", measure);
		};
	}, [title]);
	return (
		<Tooltip>
			<TooltipTrigger
				disabled={!clamped}
				render={
					<h2
						ref={ref}
						data-slot="app-shell-context-title"
						tabIndex={clamped ? 0 : undefined}
						className="line-clamp-2 min-w-0 rounded-sm px-2 py-1 text-xl leading-snug font-semibold tracking-tight text-balance wrap-anywhere outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
					>
						{title}
					</h2>
				}
			/>
			<TooltipContent side="right" className="max-w-[min(30rem,var(--available-width),calc(100vw-1rem))] text-balance">
				{title}
			</TooltipContent>
		</Tooltip>
	);
};

const toggleFade = "transition-opacity duration-(--motion-fast) ease-(--ease-out) motion-reduce:transition-none";

const NavigationPanel = ({ children, direction, reduceMotion }: { children: ReactNode; direction: number; reduceMotion: boolean }) => {
	const present = useIsPresent();
	const ref = useRef<HTMLDivElement>(null);
	useEffect(() => {
		// Never leave keyboard focus in an outgoing, inert navigation panel.
		if (!present && ref.current?.contains(document.activeElement)) {
			ref.current.parentElement?.querySelector<HTMLElement>("[data-navigation-panel]:not([inert])")?.focus();
		}
	}, [present]);
	return (
		<motion.div
			ref={ref}
			data-navigation-panel=""
			data-navigation-direction={direction}
			tabIndex={-1}
			inert={!present}
			aria-hidden={!present || undefined}
			custom={direction}
			variants={{
				enter: (travel: number) => ({
					opacity: reduceMotion ? 1 : 0,
					x: reduceMotion ? 0 : travel * 28,
				}),
				exit: (travel: number) => ({
					opacity: reduceMotion ? 1 : 0,
					x: reduceMotion ? 0 : travel * -28,
				}),
				visible: { opacity: 1, x: 0 },
			}}
			initial="enter"
			animate="visible"
			exit="exit"
			transition={{
				duration: reduceMotion ? 0 : brandMotion.slow,
				ease: brandMotion.easeOut,
			}}
			className="col-start-1 row-start-1 min-w-0 self-start outline-none"
		>
			{children}
		</motion.div>
	);
};

const useNavigationDirection = (navigationContext: AppShellProps["navigationContext"]) => {
	const [previousContext, setPreviousContext] = useState(navigationContext);
	const [direction, setDirection] = useState(1);
	if (navigationContext?.key !== previousContext?.key) {
		setPreviousContext(navigationContext);
		setDirection((navigationContext?.depth ?? 0) >= (previousContext?.depth ?? 0) ? 1 : -1);
	}
	return direction;
};

export const AppShell = ({
	labels,
	brand,
	navigation,
	children,
	showHeader = true,
	inset = false,
	header,
	actions,
	sidebarHeader,
	sidebarFooter,
	onSearch,
	searchLabel = "Search workspace",
	defaultCollapsed = false,
	collapsed: controlledCollapsed,
	onCollapsedChange,
	navigationLabel = "Main navigation",
	navigationContext,
	sidebarContext,
	className,
	contentClassName,
}: AppShellProps) => {
	const [internalCollapsed, setInternalCollapsed] = useState(defaultCollapsed);
	const [mobileOpen, setMobileOpen] = useState(false);
	// True while the rail width animates: the pointer still rests on the toggle after a click, and the
	// hover reveal (open icon) would flash before the mark settles.
	const [settling, setSettling] = useState(false);
	const collapsed = controlledCollapsed ?? internalCollapsed;
	const reduceMotion = useReducedMotion();
	const navigationDirection = useNavigationDirection(navigationContext);
	// Animate real width as in Universe: FLIP scaling distorts nested scroll areas.
	const sidebarTransition = {
		duration: reduceMotion ? 0 : brandMotion.slow,
		ease: brandMotion.easeOut,
	};
	const labelTransition = {
		duration: reduceMotion ? 0 : brandMotion.fast,
		ease: brandMotion.easeOut,
	};
	const contentId = useId();
	const sidebarId = useId();
	const closeNavigation = () => setMobileOpen(false);
	const toggleSidebar = () => {
		if (!reduceMotion) setSettling(true);
		if (controlledCollapsed === undefined) setInternalCollapsed(!collapsed);
		onCollapsedChange?.(!collapsed);
	};
	const toggleInSidebar = !showHeader || inset;
	const toggleLabel = collapsed ? (labels?.expandSidebar ?? "Expand sidebar") : (labels?.collapseSidebar ?? "Collapse sidebar");

	useEffect(() => {
		if (!onSearch) {
			return;
		}
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.defaultPrevented || event.repeat || event.altKey) {
				return;
			}
			if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
				event.preventDefault();
				onSearch();
			}
		};
		document.addEventListener("keydown", handleKeyDown);
		return () => document.removeEventListener("keydown", handleKeyDown);
	}, [onSearch]);
	useEffect(() => {
		const media = window.matchMedia("(min-width: 768px)");
		const closeOnDesktop = () => {
			if (media.matches) {
				setMobileOpen(false);
			}
		};
		media.addEventListener("change", closeOnDesktop);
		return () => media.removeEventListener("change", closeOnDesktop);
	}, []);

	const renderNavigation = (compact: boolean, mobile = false) => (
		<nav aria-label={navigationLabel} className="flex flex-col gap-1 p-2">
			{navigation.map((item) => {
				const label = (
					<>
						<span aria-hidden="true" className="flex size-5 shrink-0 items-center justify-center">
							{item.icon ?? item.label.slice(0, 1)}
						</span>
						<motion.span
							initial={false}
							animate={{ opacity: compact ? 0 : 1 }}
							transition={labelTransition}
							aria-hidden={compact || undefined}
							className="min-w-0 truncate"
						>
							{item.label}
						</motion.span>
					</>
				);
				const props = {
					"aria-current": item.active ? ("page" as const) : undefined,
					"aria-label": item.label,
					className: cn(
						"h-10 w-full justify-start gap-3 overflow-hidden px-1.5",
						item.active && "bg-background text-foreground hover:bg-background hover:text-foreground vandor-dark:hover:bg-background",
					),
					onClick: () => {
						item.onClick?.();
						if (mobile) {
							closeNavigation();
						}
					},
				};
				const control =
					item.href && !item.disabled ? (
						<Button variant="ghost" nativeButton={false} render={<a href={item.href} aria-label={item.label} />} {...props}>
							{label}
						</Button>
					) : (
						<Button variant="ghost" disabled={item.disabled} {...props}>
							{label}
						</Button>
					);
				return (
					<div key={item.id}>
						{item.renderControl ? (
							(() => {
								const custom = item.renderControl({ closeNavigation, collapsed: compact });
								if (item.tooltip === false || !isValidElement(custom)) return custom;
								return (
									<Tooltip>
										<TooltipTrigger disabled={!compact || mobile} render={custom} />
										<TooltipContent side="right">{item.label}</TooltipContent>
									</Tooltip>
								);
							})()
						) : (
							<Tooltip>
								<TooltipTrigger disabled={!compact || mobile} render={control} />
								<TooltipContent side="right">{item.label}</TooltipContent>
							</Tooltip>
						)}
						{item.dividerAfter && <div data-navigation-separator={item.id} role="separator" className="my-2 border-b border-border" />}
					</div>
				);
			})}
		</nav>
	);
	const renderPanel = (compact: boolean, mobile: boolean) => (
		<>
			{sidebarContext && (
				<div data-slot="app-shell-context" className="flex min-w-0 flex-col border-b border-border px-2 py-3">
					<div data-slot="app-shell-context-back" className="min-w-0">
						{sidebarContext.back.render({ closeNavigation, collapsed: compact })}
					</div>
					{/* Height animates with the rail so the navigation below slides instead of jumping. */}
					<motion.div
						initial={false}
						animate={{ height: compact ? 0 : "auto", opacity: compact ? 0 : 1 }}
						transition={sidebarTransition}
						inert={compact || undefined}
						className="overflow-hidden"
					>
						<div className="flex min-w-0 flex-col gap-2 pt-2">
							{/* Only the title wraps, so only it keeps the expanded width (15rem rail minus its px-2 gutters)
							    while collapsed or animating: the measured auto height is then the final one. The single-row
							    mode switch keeps following the real width, so nothing adjusts when the animation ends. */}
							<div className={cn("min-w-0", !mobile && (compact || settling) && "w-[calc(15rem-1rem)]")}>
								<ContextTitle title={sidebarContext.title} />
							</div>
							{sidebarContext.meta ? <div data-slot="app-shell-context-meta">{sidebarContext.meta}</div> : null}
						</div>
					</motion.div>
				</div>
			)}
			{renderNavigation(compact, mobile)}
		</>
	);
	const renderContextNavigation = (compact: boolean, mobile = false) =>
		navigationContext ? (
			<div className="grid min-w-0 overflow-x-clip">
				<AnimatePresence initial={false} custom={navigationDirection}>
					<NavigationPanel key={navigationContext.key} direction={navigationDirection} reduceMotion={!!reduceMotion}>
						{renderPanel(compact, mobile)}
					</NavigationPanel>
				</AnimatePresence>
			</div>
		) : (
			renderPanel(compact, mobile)
		);

	const renderBrand = (compact: boolean, inline = false, className?: string) => {
		const content = (
			<>
				<span aria-hidden="true" className="flex size-6 shrink-0 items-center justify-center font-semibold">
					{brand.icon ?? brand.name.slice(0, 1)}
				</span>
				<motion.span
					initial={false}
					animate={{ opacity: compact ? 0 : 1 }}
					transition={labelTransition}
					aria-hidden={compact || brand.label !== undefined || undefined}
					className={cn("min-w-0 truncate", brand.label === undefined && zenithWordmarkClassName)}
				>
					{brand.label ?? brand.name}
				</motion.span>
				{(compact || brand.label !== undefined) && <span className="sr-only">{brand.name}</span>}
			</>
		);
		const styles = cn(
			"flex min-w-0 items-center gap-3 overflow-hidden",
			inline ? "flex-1 ps-1 pe-2" : "h-16 shrink-0 border-b border-border px-5",
			className,
		);
		return brand.href ? (
			<a href={brand.href} aria-label={brand.name} className={cn(styles, "focus-visible:outline-2 focus-visible:outline-ring")}>
				{content}
			</a>
		) : (
			<div className={styles}>{content}</div>
		);
	};
	const renderSidebarHeader = (compact: boolean, mobile = false) =>
		toggleInSidebar && !mobile ? (
			<div data-slot="app-shell-sidebar-header" className={cn("relative flex shrink-0 items-center px-2", inset ? "h-14" : "h-16")}>
				{/* The mark never leaves or moves: it is centred on the 24px axis shared by every sidebar icon, which
				    is also the 48px rail's centre. Only the label fades. The toggle hides at once on collapse and,
				    at the rail, sits exactly over the mark, revealing the expand icon on hover. */}
				<div
					inert={compact || undefined}
					className={cn(
						"flex min-w-0 flex-1 items-center overflow-hidden transition-[padding] duration-(--motion-slow) ease-(--ease-out) motion-reduce:transition-none",
						compact ? "pe-0" : onSearch && inset ? "pe-[4.5rem]" : "pe-10",
					)}
				>
					{typeof sidebarHeader === "function" ? (
						// A render function shapes its own rail state (e.g. an avatar that stays on the icon axis), so it
						// is not faded; on the rail the toggle covers it and reveals the expand icon on hover.
						<div className="flex min-w-0 flex-1 items-center">{sidebarHeader({ closeNavigation, collapsed: compact })}</div>
					) : sidebarHeader ? (
						<motion.div
							initial={false}
							animate={{ opacity: compact ? 0 : 1 }}
							transition={labelTransition}
							className="flex min-w-0 flex-1 items-center"
						>
							{sidebarHeader}
						</motion.div>
					) : (
						renderBrand(compact, true)
					)}
				</div>
				{onSearch && inset && !compact && (
					<Tooltip>
						<TooltipTrigger
							render={
								<Button
									variant="ghost"
									size="icon"
									className="absolute top-1/2 right-10 size-8 -translate-y-1/2"
									aria-label={searchLabel}
									onClick={onSearch}
								>
									<SearchIcon aria-hidden="true" />
								</Button>
							}
						/>
						<TooltipContent side="bottom">
							{searchLabel} <kbd className="ms-1 font-sans opacity-70">⌘K</kbd>
						</TooltipContent>
					</Tooltip>
				)}
				<Tooltip>
					<TooltipTrigger
						render={
							<Button
								variant="ghost"
								size="icon"
								className={cn(
									"group/sidebar-toggle absolute top-1/2 size-8 -translate-y-1/2",
									compact ? "left-2 hover:bg-accent focus-visible:bg-accent vandor-dark:hover:bg-accent" : "right-2",
								)}
								aria-label={toggleLabel}
								aria-expanded={!collapsed}
								aria-controls={sidebarId}
								onClick={toggleSidebar}
							>
								<span data-slot="app-shell-sidebar-toggle-icon" className="relative flex size-4 items-center justify-center">
									{/* No transition into compact: the close icon vanishes at once instead of riding the edge. */}
									<PanelLeftCloseIcon aria-hidden="true" className={cn("absolute", compact ? "opacity-0" : toggleFade)} />
									<PanelLeftOpenIcon
										aria-hidden="true"
										className={cn(
											toggleFade,
											"absolute opacity-0",
											compact && !settling && "group-hover/sidebar-toggle:opacity-100 group-focus-visible/sidebar-toggle:opacity-100",
										)}
									/>
								</span>
							</Button>
						}
					/>
					<TooltipContent side="right" role="tooltip">
						{toggleLabel}
					</TooltipContent>
				</Tooltip>
			</div>
		) : sidebarHeader ? (
			<div className="flex h-16 shrink-0 items-center border-b border-border p-2">
				{typeof sidebarHeader === "function" ? sidebarHeader({ closeNavigation, collapsed: compact }) : sidebarHeader}
			</div>
		) : (
			renderBrand(compact)
		);
	const renderFooter = (compact: boolean) =>
		sidebarFooter && (
			<div className="mt-auto shrink-0 border-t border-border p-2">
				{typeof sidebarFooter === "function" ? sidebarFooter({ collapsed: compact, closeNavigation }) : sidebarFooter}
			</div>
		);
	const mobileNavigation = (
		<Drawer
			title={sidebarHeader ? navigationLabel : brand.name}
			open={mobileOpen}
			onOpenChange={setMobileOpen}
			swipeDirection={showHeader ? "left" : "right"}
			contentProps={{ closeButtonLabel: labels?.closeNavigation ?? "Close navigation", className: "bg-sidebar" }}
			showSwipeHandle={false}
			trigger={
				<Button
					variant={showHeader ? "ghost" : "outline"}
					size="icon"
					className={cn(
						"shrink-0 md:hidden",
						!showHeader &&
							"fixed top-[max(1rem,env(safe-area-inset-top))] right-[max(1rem,env(safe-area-inset-right))] z-20 size-10 bg-background shadow-sm sm:right-[max(1.5rem,env(safe-area-inset-right))]",
					)}
					aria-label={labels?.openNavigation ?? "Open navigation"}
				>
					<MenuIcon aria-hidden="true" />
				</Button>
			}
		>
			<DrawerBody className="flex min-h-0 flex-1 flex-col p-0">
				{sidebarHeader && renderSidebarHeader(false, true)}
				<div className="min-h-0 flex-1 overflow-x-hidden overflow-y-auto [scrollbar-width:thin]">
					{renderContextNavigation(false, true)}
				</div>
				{renderFooter(false)}
			</DrawerBody>
		</Drawer>
	);

	return (
		<TooltipProvider delay={250}>
			<div
				data-slot="app-shell"
				data-collapsed={collapsed}
				data-inset={inset || undefined}
				className={cn(
					"flex min-h-svh overflow-x-clip text-foreground",
					inset ? "bg-sidebar md:h-svh md:overflow-hidden" : "bg-background",
					className,
				)}
			>
				<a
					href={`#${contentId}`}
					className="sr-only z-50 rounded-md bg-background p-3 focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
				>
					{labels?.skipToContent ?? "Skip to content"}
				</a>
				<motion.aside
					id={sidebarId}
					data-slot="app-shell-sidebar"
					aria-label={labels?.sidebar ?? "Sidebar"}
					initial={false}
					animate={{ width: collapsed ? "3rem" : "15rem" }}
					transition={sidebarTransition}
					onAnimationComplete={() => setSettling(false)}
					className={cn(
						"sticky top-0 hidden h-svh shrink-0 flex-col overflow-hidden bg-sidebar md:flex",
						!inset && "border-r border-border",
					)}
				>
					{renderSidebarHeader(collapsed)}
					{/* Thin themed scrollbar when expanded; none on the 48px rail, where a classic scrollbar would push the
					    icons off their centre axis (wheel and touch scrolling still work). */}
					<div
						className={cn(
							"min-h-0 flex-1 overflow-x-hidden overflow-y-auto",
							collapsed ? "[scrollbar-width:none] [&::-webkit-scrollbar]:hidden" : "[scrollbar-width:thin]",
						)}
					>
						{renderContextNavigation(collapsed)}
					</div>
					{renderFooter(collapsed)}
				</motion.aside>
				<div
					data-slot="app-shell-panel"
					className={cn(
						"flex min-w-0 flex-1 flex-col",
						// The panel's own surface: everything inside that paints `bg-background` follows it.
						inset &&
							"bg-background [--background:var(--card)] md:my-2 md:me-2 md:min-h-0 md:overflow-hidden md:rounded-xl md:border md:border-border md:shadow-xs",
					)}
				>
					{showHeader ? (
						<header
							data-slot="app-shell-header"
							className={cn(
								"sticky top-0 z-20 flex shrink-0 items-center gap-3 border-b border-border bg-background px-4 sm:px-6",
								inset ? "h-12 md:px-4" : "h-16",
							)}
						>
							{mobileNavigation}
							<Button
								variant="ghost"
								size="icon"
								className={cn("hidden", !inset && "md:inline-flex")}
								aria-label={collapsed ? (labels?.expandSidebar ?? "Expand sidebar") : (labels?.collapseSidebar ?? "Collapse sidebar")}
								aria-expanded={!collapsed}
								aria-controls={sidebarId}
								onClick={toggleSidebar}
							>
								{collapsed ? <PanelLeftOpenIcon aria-hidden="true" /> : <PanelLeftCloseIcon aria-hidden="true" />}
							</Button>
							{onSearch && inset && (
								<Button variant="ghost" size="icon" className="order-last md:hidden" aria-label={searchLabel} onClick={onSearch}>
									<SearchIcon aria-hidden="true" />
								</Button>
							)}
							{onSearch && !inset && (
								<Button variant="outline" onClick={onSearch} className="min-w-0 flex-1 justify-start text-muted-foreground sm:max-w-sm">
									<SearchIcon aria-hidden="true" />
									<span className="truncate">{searchLabel}</span>
									<kbd className="ml-auto hidden text-xs sm:inline">⌘ / Ctrl K</kbd>
								</Button>
							)}
							{header}
							{actions && <div className="ml-auto flex shrink-0 items-center gap-2">{actions}</div>}
						</header>
					) : (
						<div data-slot="app-shell-mobile-controls" className="h-[calc(max(1rem,env(safe-area-inset-top))+3.5rem)] shrink-0 md:hidden">
							<div
								data-slot="app-shell-mobile-brand"
								className="fixed top-0 left-[max(1rem,env(safe-area-inset-left))] z-20 isolate flex h-[calc(max(1rem,env(safe-area-inset-top))+2.5rem)] max-w-[calc(100%-8rem)] items-center px-3 pt-[max(1rem,env(safe-area-inset-top))] sm:left-[max(1.5rem,env(safe-area-inset-left))]"
							>
								<svg
									aria-hidden="true"
									viewBox="0 0 160 56"
									preserveAspectRatio="none"
									className="pointer-events-none absolute top-0 -left-4 -z-10 h-full w-[calc(100%+2rem)] overflow-visible drop-shadow-sm"
								>
									<path
										d="M0 0H160V1C151 1 143 8 143 18V35C143 46 134 55 123 55H37C26 55 17 46 17 35V18C17 8 9 1 0 1Z"
										className="fill-background"
									/>
									<path
										d="M0 1C9 1 17 8 17 18V35C17 46 26 55 37 55H123C134 55 143 46 143 35V18C143 8 151 1 160 1"
										fill="none"
										strokeWidth="1"
										vectorEffect="non-scaling-stroke"
										className="stroke-border vandor-dark:stroke-input"
									/>
								</svg>
								{renderBrand(false, true, "gap-1.5")}
							</div>
							{mobileNavigation}
						</div>
					)}
					<main
						id={contentId}
						data-slot="app-shell-main"
						data-scroll-restoration-id={inset ? "app-shell-main" : undefined}
						tabIndex={-1}
						// `relative` contains absolute descendants (e.g. `sr-only` chart tables): otherwise they resolve
						// against the page, and the locked window grows a second scrollbar.
						className={cn("min-w-0 flex-1 outline-none", inset && "relative md:min-h-0 md:overflow-y-auto", contentClassName)}
					>
						{children}
					</main>
				</div>
			</div>
		</TooltipProvider>
	);
};
