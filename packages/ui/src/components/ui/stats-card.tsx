"use client";

// Adapted from Zenith's StatsCard. Uses local supporting styles, not website UI.
import { motion as brandMotion } from "@repo/brand/motion";
import { cn } from "cn";
import type { LucideIcon } from "lucide-react";
import { animate, motion, useMotionValue, useReducedMotion, useTransform } from "motion/react";
import { useEffect, useRef, useState } from "react";
import type { ReactNode } from "react";

import { StatsCardBadge } from "./stats-card-badge";
import type { StatsCardBadgeVariant } from "./stats-card-badge";

export type StatsCardValue =
	| {
			target: number | null;
			format?: (value: number) => string;
			display?: never;
	  }
	| { display: ReactNode; target?: never; format?: never };

export interface StatsCardItem {
	key: string;
	title: string;
	value: StatsCardValue;
	icon: LucideIcon;
	caption?: ReactNode;
	badge?: {
		label: ReactNode;
		variant?: StatsCardBadgeVariant;
		className?: string;
	};
}

export interface StatsCardProps {
	ariaLabel: string;
	items: readonly StatsCardItem[];
	className?: string;
}

const defaultNumberFormat = new Intl.NumberFormat("id-ID", {
	maximumFractionDigits: 0,
}).format;

const AnimatedStatsValue = ({ value }: { value: Extract<StatsCardValue, { target: number | null }> }) => {
	const target = value.target !== null && Number.isFinite(value.target) ? value.target : null;
	const format = value.format ?? defaultNumberFormat;
	const formatRef = useRef(format);
	formatRef.current = format;
	const motionValue = useMotionValue(0);
	const formattedValue = useTransform(motionValue, (current) => formatRef.current(current));
	const shouldReduceMotion = useReducedMotion();
	const [hasMounted, setHasMounted] = useState(false);
	const hasStarted = useRef(false);
	const finalValue = target === null ? "—" : format(target);

	useEffect(() => {
		if (target === null) {
			motionValue.set(0);
			// The mount pass is over even without a value, so a later target is set directly.
			hasStarted.current = true;
			setHasMounted(true);
			return;
		}
		if (shouldReduceMotion) {
			motionValue.set(target);
			hasStarted.current = true;
			setHasMounted(true);
			return;
		}
		if (hasStarted.current) {
			// Count-up runs once on mount; later data changes never animate.
			motionValue.set(target);
			return;
		}
		motionValue.set(0);
		hasStarted.current = true;
		const controls = animate(motionValue, target, {
			duration: brandMotion.slow,
			ease: brandMotion.easeOut,
		});
		setHasMounted(true);
		return () => controls.stop();
	}, [motionValue, shouldReduceMotion, target]);

	if (!hasMounted || shouldReduceMotion || target === null) {
		return <span>{finalValue}</span>;
	}
	return (
		<>
			<span className="sr-only">{finalValue}</span>
			<motion.span aria-hidden="true">{formattedValue}</motion.span>
		</>
	);
};

const StatsValue = ({ value }: { value: StatsCardValue }) =>
	"display" in value ? <span>{value.display}</span> : <AnimatedStatsValue value={value} />;

const getItemLayout = (count: number, maxColumns: number, index: number) => {
	const rows = Math.ceil(count / maxColumns);
	const shortRowSize = Math.floor(count / rows);
	const longRows = count % rows;
	const longRowItems = longRows * (shortRowSize + 1);
	// Distribute spare items to the first rows, keeping row sizes within one item.
	const columns = index < longRowItems ? shortRowSize + 1 : shortRowSize;
	const row = index < longRowItems ? Math.floor(index / columns) : longRows + Math.floor((index - longRowItems) / columns);
	const column = index < longRowItems ? index % columns : (index - longRowItems) % columns;
	return { columns, lastRow: row === rows - 1, lastColumn: column === columns - 1 };
};

export const StatsCard = ({ ariaLabel, items, className }: StatsCardProps) => {
	return (
		<div
			data-slot="stats-card"
			role="group"
			aria-label={ariaLabel}
			className={cn("w-full min-w-0 overflow-hidden rounded-xl border border-border bg-card text-card-foreground", className)}
		>
			<dl className="grid grid-cols-1 sm:grid-cols-12">
				{items.map((item, index) => {
					const Icon = item.icon;
					const hasCaption = item.caption !== null && item.caption !== undefined;
					const small = getItemLayout(items.length, items.length === 3 ? 3 : 2, index);
					const large = getItemLayout(items.length, 4, index);
					return (
						<div
							key={item.key}
							data-slot="stats-card-item"
							className={cn(
								"min-w-0 border-border",
								index < items.length - 1 && "border-b",
								small.columns === 1 ? "sm:col-span-12" : small.columns === 2 ? "sm:col-span-6" : "sm:col-span-4",
								large.columns === 1
									? "lg:col-span-12"
									: large.columns === 2
										? "lg:col-span-6"
										: large.columns === 3
											? "lg:col-span-4"
											: "lg:col-span-3",
								small.lastRow ? "sm:border-b-0" : "sm:border-b",
								small.lastColumn ? "sm:border-e-0" : "sm:border-e",
								large.lastRow ? "lg:border-b-0" : "lg:border-b",
								large.lastColumn ? "lg:border-e-0" : "lg:border-e",
							)}
						>
							<div className="flex h-full items-start justify-between gap-4 p-5">
								<div className="flex min-w-0 flex-1 flex-col gap-3">
									<dt className="text-sm font-medium wrap-anywhere">{item.title}</dt>
									<dd className="min-w-0">
										<div className="text-2xl leading-tight font-semibold tracking-tight tabular-nums wrap-anywhere">
											<StatsValue value={item.value} />
										</div>
										{(hasCaption || item.badge) && (
											<div className="mt-1 flex flex-wrap items-center gap-2">
												{hasCaption && <div className="text-xs text-muted-foreground wrap-anywhere">{item.caption}</div>}
												{item.badge && (
													<StatsCardBadge className={item.badge.className} variant={item.badge.variant}>
														{item.badge.label}
													</StatsCardBadge>
												)}
											</div>
										)}
									</dd>
								</div>
								<div className="flex size-10 shrink-0 items-center justify-center rounded-full border border-border text-muted-foreground [&_svg]:size-4">
									<Icon aria-hidden="true" />
								</div>
							</div>
						</div>
					);
				})}
			</dl>
		</div>
	);
};
