"use client";

import { cn } from "cn";
import { useReducedMotion } from "motion/react";
import { useId } from "react";
import type { ReactNode } from "react";
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

/*
 * Thin Recharts wrappers in the design-system voice: ink-colored marks, a recessive grid, one axis,
 * a hover tooltip, and an sr-only table so the data never depends on seeing the plot.
 */

const ink = "var(--primary)";
const muted = "var(--muted-foreground)";
const grid = "var(--border)";

export interface TrendPoint {
	/** Axis and tooltip label, already formatted for people. */
	label: string;
	value: number;
}

export interface BarPoint {
	key: string;
	label: string;
	value: number;
	/** Any CSS color; defaults to ink. Pair a semantic color with its label, never color alone. */
	color?: string;
}

type ChartTooltipProps = {
	active?: boolean;
	payload?: { value?: number; payload?: { label?: string; color?: string } }[];
	format: (value: number) => string;
};

// Above every sibling (a donut's center label paints later in the DOM), never catching the pointer, and snapping
// with the cursor instead of sliding between marks.
const tooltipProps = { wrapperStyle: { zIndex: 50, pointerEvents: "none" as const }, isAnimationActive: false, offset: 14 };

const ChartTooltip = ({ active, payload, format }: ChartTooltipProps) => {
	const item = payload?.[0];
	if (!active || !item || item.value === undefined) return null;
	const swatch = item.payload?.color;
	return (
		<div className="min-w-32 rounded-md bg-popover px-3 py-2 text-xs text-popover-foreground shadow-md ring-1 ring-foreground/10">
			<p className="flex items-center gap-1.5 text-muted-foreground">
				{swatch && <span aria-hidden="true" className="size-2 shrink-0 rounded-sm" style={{ background: swatch }} />}
				{item.payload?.label}
			</p>
			<p className="mt-0.5 text-sm font-semibold tabular-nums">{format(item.value)}</p>
		</div>
	);
};

const DataTable = ({
	caption,
	rows,
	format,
}: {
	caption: string;
	rows: readonly { label: string; value: number }[];
	format: (v: number) => string;
}) => (
	<table className="sr-only">
		<caption>{caption}</caption>
		<tbody>
			{rows.map((row) => (
				<tr key={row.label}>
					<th scope="row">{row.label}</th>
					<td>{format(row.value)}</td>
				</tr>
			))}
		</tbody>
	</table>
);

export interface AreaTrendProps {
	data: readonly TrendPoint[];
	/** Accessible name of the figure; also the table caption. */
	label: string;
	formatValue: (value: number) => string;
	/** Compact y-axis tick text; defaults to formatValue. */
	formatTick?: (value: number) => string;
	/** Mark color; ink by default. */
	color?: string;
	height?: number;
	className?: string;
}

export const AreaTrend = ({ data, label, formatValue, formatTick, color = ink, height = 220, className }: AreaTrendProps) => {
	const gradient = useId();
	const reduceMotion = useReducedMotion();
	return (
		<figure className={cn("min-w-0", className)} aria-label={label} style={{ height }}>
			<ResponsiveContainer width="100%" height="100%">
				<AreaChart data={data as TrendPoint[]} margin={{ top: 8, right: 8, bottom: 0, left: 0 }} accessibilityLayer={false}>
					<defs>
						<linearGradient id={gradient} x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stopColor={color} stopOpacity={0.18} />
							<stop offset="100%" stopColor={color} stopOpacity={0} />
						</linearGradient>
					</defs>
					<CartesianGrid vertical={false} stroke={grid} strokeDasharray="3 3" />
					<XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: muted, fontSize: 11 }} minTickGap={24} />
					<YAxis
						tickLine={false}
						axisLine={false}
						width={48}
						tick={{ fill: muted, fontSize: 11 }}
						tickFormatter={formatTick ?? formatValue}
						tickCount={4}
					/>
					<Tooltip {...tooltipProps} cursor={{ stroke: muted, strokeDasharray: "3 3" }} content={<ChartTooltip format={formatValue} />} />
					<Area
						type="monotone"
						dataKey="value"
						stroke={color}
						strokeWidth={2}
						fill={`url(#${gradient})`}
						isAnimationActive={!reduceMotion}
						activeDot={{ r: 4, stroke: "var(--background)", strokeWidth: 2 }}
						dot={false}
					/>
				</AreaChart>
			</ResponsiveContainer>
			<DataTable caption={label} rows={data} format={formatValue} />
		</figure>
	);
};

export interface BarListProps {
	data: readonly BarPoint[];
	label: string;
	formatValue: (value: number) => string;
	className?: string;
	/** Rendered when `data` is empty. */
	empty?: ReactNode;
}

const rowHeight = 36;

export const BarList = ({ data, label, formatValue, className, empty }: BarListProps) => {
	const reduceMotion = useReducedMotion();
	if (data.length === 0) return <>{empty}</>;
	return (
		<figure className={cn("min-w-0", className)} aria-label={label} style={{ height: data.length * rowHeight + 8 }}>
			<ResponsiveContainer width="100%" height="100%">
				<BarChart data={data as BarPoint[]} layout="vertical" margin={{ top: 0, right: 40, bottom: 0, left: 0 }} barCategoryGap={10}>
					<XAxis type="number" hide domain={[0, "dataMax"]} />
					<YAxis
						type="category"
						dataKey="label"
						tickLine={false}
						axisLine={false}
						width={112}
						tick={{ fill: "var(--foreground)", fontSize: 12 }}
					/>
					<Tooltip {...tooltipProps} cursor={{ fill: "var(--accent)", opacity: 0.5 }} content={<ChartTooltip format={formatValue} />} />
					<Bar
						dataKey="value"
						radius={[0, 4, 4, 0]}
						isAnimationActive={!reduceMotion}
						label={{
							position: "right",
							fill: "var(--foreground)",
							fontSize: 12,
							formatter: (value: unknown) => formatValue(Number(value)),
						}}
					>
						{data.map((point) => (
							<Cell key={point.key} fill={point.color ?? ink} />
						))}
					</Bar>
				</BarChart>
			</ResponsiveContainer>
			<DataTable caption={label} rows={data} format={formatValue} />
		</figure>
	);
};

/** Categorical hues from the brand tokens, in their validated order. Fold extra slices into a neutral "Lainnya" rather than cycling. */
export const chartRamp = ["var(--viz-1)", "var(--viz-2)", "var(--viz-3)", "var(--viz-4)", "var(--viz-5)"] as const;

export interface DonutProps {
	data: readonly BarPoint[];
	label: string;
	formatValue: (value: number) => string;
	/** Big text in the hole, e.g. the total. */
	center?: { value: string; caption: string };
	className?: string;
}

export const Donut = ({ data, label, formatValue, center, className }: DonutProps) => {
	const reduceMotion = useReducedMotion();
	const total = data.reduce((sum, point) => sum + point.value, 0);
	const colored = data.map((point, index) => ({ ...point, color: point.color ?? chartRamp[index % chartRamp.length] }));
	return (
		<figure className={cn("flex min-w-0 flex-wrap items-center gap-x-6 gap-y-4", className)} aria-label={label}>
			<div className="relative size-40 shrink-0">
				<ResponsiveContainer width="100%" height="100%">
					<PieChart>
						<Tooltip {...tooltipProps} content={<ChartTooltip format={formatValue} />} />
						<Pie
							data={colored}
							dataKey="value"
							nameKey="label"
							innerRadius="64%"
							outerRadius="100%"
							paddingAngle={colored.length > 1 ? 2 : 0}
							stroke="var(--background)"
							strokeWidth={2}
							isAnimationActive={!reduceMotion}
						>
							{colored.map((point) => (
								<Cell key={point.key} fill={point.color} />
							))}
						</Pie>
					</PieChart>
				</ResponsiveContainer>
				{center && (
					<div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center text-center">
						<span className="text-xl font-semibold tabular-nums">{center.value}</span>
						<span className="text-xs text-muted-foreground">{center.caption}</span>
					</div>
				)}
			</div>
			<ul className="flex min-w-40 flex-1 flex-col gap-2 text-sm">
				{colored.map((point) => (
					<li key={point.key} className="flex items-center gap-2">
						<span aria-hidden="true" className="size-2.5 shrink-0 rounded-sm" style={{ background: point.color }} />
						<span className="min-w-0 flex-1 truncate">{point.label}</span>
						<span className="tabular-nums text-muted-foreground">
							{formatValue(point.value)} · {total > 0 ? Math.round((point.value / total) * 100) : 0}%
						</span>
					</li>
				))}
			</ul>
			<DataTable caption={label} rows={data} format={formatValue} />
		</figure>
	);
};

export interface StackSeries {
	key: string;
	label: string;
	color: string;
}

export interface StackedColumnsProps {
	/** One row per x label; each series key holds that row's value. */
	data: readonly ({ label: string } & Record<string, number | string>)[];
	series: readonly StackSeries[];
	label: string;
	formatValue: (value: number) => string;
	height?: number;
	className?: string;
}

type StackTooltipProps = {
	active?: boolean;
	label?: string;
	payload?: { dataKey?: string; value?: number }[];
	series: readonly StackSeries[];
	format: (value: number) => string;
};

const StackTooltip = ({ active, label, payload, series, format }: StackTooltipProps) => {
	if (!active || !payload?.length) return null;
	return (
		<div className="min-w-36 rounded-md bg-popover px-3 py-2 text-xs text-popover-foreground shadow-md ring-1 ring-foreground/10">
			<p className="mb-1 text-muted-foreground">{label}</p>
			{series.map((item) => (
				<p key={item.key} className="flex items-center gap-2 tabular-nums">
					<span aria-hidden="true" className="size-2 rounded-sm" style={{ background: item.color }} />
					<span className="flex-1">{item.label}</span>
					<span className="font-medium">{format(Number(payload.find((entry) => entry.dataKey === item.key)?.value ?? 0))}</span>
				</p>
			))}
		</div>
	);
};

export const StackedColumns = ({ data, series, label, formatValue, height = 220, className }: StackedColumnsProps) => {
	const reduceMotion = useReducedMotion();
	return (
		<figure className={cn("flex min-w-0 flex-col gap-3", className)} aria-label={label}>
			<ul className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
				{series.map((item) => (
					<li key={item.key} className="flex items-center gap-1.5">
						<span aria-hidden="true" className="size-2.5 rounded-sm" style={{ background: item.color }} />
						{item.label}
					</li>
				))}
			</ul>
			<div style={{ height }}>
				<ResponsiveContainer width="100%" height="100%">
					<BarChart data={[...data]} margin={{ top: 4, right: 8, bottom: 0, left: 0 }} barCategoryGap="20%">
						<CartesianGrid vertical={false} stroke={grid} strokeDasharray="3 3" />
						<XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: muted, fontSize: 11 }} minTickGap={24} />
						<YAxis tickLine={false} axisLine={false} width={36} tick={{ fill: muted, fontSize: 11 }} tickCount={4} />
						<Tooltip
							{...tooltipProps}
							cursor={{ fill: "var(--accent)", opacity: 0.5 }}
							content={<StackTooltip series={series} format={formatValue} />}
						/>
						{series.map((item, index) => (
							<Bar
								key={item.key}
								dataKey={item.key}
								stackId="stack"
								fill={item.color}
								stroke="var(--background)"
								strokeWidth={1}
								radius={index === series.length - 1 ? [3, 3, 0, 0] : 0}
								isAnimationActive={!reduceMotion}
							/>
						))}
					</BarChart>
				</ResponsiveContainer>
			</div>
			<table className="sr-only">
				<caption>{label}</caption>
				<tbody>
					{data.map((row) => (
						<tr key={row.label}>
							<th scope="row">{row.label}</th>
							{series.map((item) => (
								<td key={item.key}>
									{item.label}: {formatValue(Number(row[item.key] ?? 0))}
								</td>
							))}
						</tr>
					))}
				</tbody>
			</table>
		</figure>
	);
};
