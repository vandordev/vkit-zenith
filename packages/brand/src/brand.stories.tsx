import type { Meta, StoryObj } from "@storybook/react";
import { motion } from "./motion";
import { zenithMark } from "./logo-mark";
import { brand } from "./index";
import "./tokens.css";

/** Token-colored mark (same geometry as ZenithMark; brand cannot import @repo/ui). */
function ThemedMark({ size = 180 }: { size?: number }) {
	return (
		<svg viewBox={zenithMark.viewBox} width={size} height={size} role="img" aria-label={`${brand.name} logo`}>
			<g style={{ fill: "var(--primary)" }}>
				{zenithMark.paths.map((d) => (
					<path key={d} d={d} />
				))}
			</g>
		</svg>
	);
}

const meta = {
	title: "Brand/Identity",
	parameters: { layout: "centered" },
} satisfies Meta;
export default meta;
type Story = StoryObj<typeof meta>;

export const Identity: Story = {
	render: () => (
		<div className="flex flex-col items-center gap-6">
			<div className="p-6">
				<ThemedMark />
			</div>
			<p className="text-xl font-semibold">{brand.name}</p>
			<p className="max-w-sm text-center text-sm text-muted-foreground">Zenith logo, ink-on-cream color tokens, fonts and motion.</p>
		</div>
	),
};

const radiusRoles = [
	{ name: "sm", className: "rounded-sm" },
	{ name: "md", className: "rounded-md" },
	{ name: "lg", className: "rounded-lg" },
	{ name: "xl", className: "rounded-xl" },
	{ name: "2xl", className: "rounded-2xl" },
	{ name: "3xl", className: "rounded-3xl" },
	{ name: "4xl", className: "rounded-4xl" },
] as const;

export const RadiusTokens: Story = {
	render: () => (
		<div className="flex max-w-2xl flex-col gap-6">
			<p className="text-sm text-muted-foreground">
				Shared radius scale. Base: 0.8rem. Circular shapes and component-specific geometry are preserved.
			</p>
			<div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
				{radiusRoles.map(({ name, className }) => (
					<div key={name} className="flex flex-col gap-2">
						<div aria-hidden="true" className={`h-20 w-full border border-border bg-muted ${className}`} />
						<code className="text-sm">rounded-{name}</code>
					</div>
				))}
			</div>
		</div>
	),
};

const colorRoles = [
	"background",
	"foreground",
	"primary",
	"primary-foreground",
	"muted",
	"muted-foreground",
	"secondary",
	"secondary-foreground",
	"accent",
	"accent-foreground",
	"destructive",
	"destructive-foreground",
	"destructive-text",
	"success",
	"success-foreground",
	"success-text",
	"info",
	"info-foreground",
	"info-text",
	"warning",
	"warning-foreground",
	"warning-text",
	"test",
	"test-foreground",
	"test-text",
	"popover",
	"popover-foreground",
	"card",
	"card-foreground",
	"input",
	"border",
	"ring",
	"chart-1",
	"chart-2",
	"chart-3",
	"chart-4",
	"chart-5",
	"sidebar",
	"sidebar-foreground",
	"sidebar-primary",
	"sidebar-primary-foreground",
	"sidebar-accent",
	"sidebar-accent-foreground",
	"sidebar-border",
	"sidebar-ring",
] as const;

export const ColorTokens: Story = {
	render: () => (
		<div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
			{colorRoles.map((role) => (
				<div key={role} className="flex items-center gap-3">
					<span
						aria-hidden="true"
						className="size-10 shrink-0 rounded-md border border-border"
						style={{ backgroundColor: `var(--${role})` }}
					/>
					<code className="text-sm">--{role}</code>
				</div>
			))}
		</div>
	),
};

export const DarkIdentity: Story = {
	render: () => (
		<div className="dark w-full rounded-lg bg-background p-8 text-foreground">
			<div className="flex flex-col items-center gap-6">
				<ThemedMark />
				<p className="text-xl font-semibold">{brand.name}</p>
				<p className="text-sm text-muted-foreground">Zenith dark palette</p>
			</div>
		</div>
	),
};

export const DarkColorTokens: Story = {
	...ColorTokens,
	decorators: [
		(Story) => (
			<div className="dark rounded-lg bg-background p-6 text-foreground">
				<Story />
			</div>
		),
	],
};

const status = ["success", "warning", "info", "destructive", "test"] as const;

export const StatusOnSurface: Story = {
	render: () => (
		<div className="flex flex-wrap gap-3">
			{status.map((role) => (
				<span
					key={role}
					className="rounded-md px-3 py-1.5 text-sm font-medium"
					style={{ color: `var(--${role}-text)`, backgroundColor: `color-mix(in oklab, var(--${role}-text) 10%, var(--card))` }}
				>
					{role}
				</span>
			))}
		</div>
	),
};

export const StatusOnSurfaceDark: Story = {
	...StatusOnSurface,
	decorators: [
		(Story) => (
			<div className="dark rounded-lg bg-card p-6 text-foreground">
				<Story />
			</div>
		),
	],
};

const specimen = [
	{ family: "font-sans", label: "Plus Jakarta Sans: teks antarmuka", sample: "Pembayaran berhasil, Rp1.250.000" },
	{ family: "font-mono", label: "JetBrains Mono: ID, kode, nomor VA", sample: "va_8F2K-0019 · 8808 1234 5678" },
] as const;

export const Typography: Story = {
	render: () => (
		<div className="flex max-w-2xl flex-col gap-8">
			{specimen.map(({ family, label, sample }) => (
				<div key={family} className="flex flex-col gap-2">
					<code className="text-xs text-muted-foreground">{label}</code>
					<p className={`${family} text-3xl`}>{sample}</p>
					<p className={`${family} text-sm`}>0123456789 ABCDEFGHIJ abcdefghij</p>
				</div>
			))}
			<div className="flex flex-col gap-2">
				<code className="text-xs text-muted-foreground">Tabular figures (tnum) on Plus Jakarta Sans</code>
				<p className="text-xl [font-variant-numeric:tabular-nums]">
					1111111
					<br />
					8888888
				</p>
			</div>
		</div>
	),
};

export const MotionTokens: Story = {
	render: () => (
		<div className="grid grid-cols-[auto_auto_1fr] gap-x-6 gap-y-1 text-sm">
			{(
				[
					["--motion-fast", `${motion.fast * 1000}ms`, "hover, press"],
					["--motion-base", `${motion.base * 1000}ms`, "status, toast, content in"],
					["--motion-slow", `${motion.slow * 1000}ms`, "panel transitions"],
					["--ease-out", `cubic-bezier(${motion.easeOut.join(", ")})`, "almost everything"],
					["--ease-in-out", `cubic-bezier(${motion.easeInOut.join(", ")})`, "pulses only"],
				] as const
			).map(([name, value, use]) => (
				<div key={name} className="contents">
					<code>{name}</code>
					<span>{value}</span>
					<span className="text-muted-foreground">{use}</span>
				</div>
			))}
		</div>
	),
};
