import type { Meta, StoryObj } from "@storybook/react";
import { Activity, CircleCheck, Pause, Wallet } from "lucide-react";
import { useState } from "react";
import { expect, userEvent, waitFor, within } from "storybook/test";

import { StatsCard } from "./stats-card";
import type { StatsCardProps } from "./stats-card";

const items: StatsCardProps["items"] = [
	{
		badge: { label: "Active" },
		caption: "Ready to receive events",
		icon: Activity,
		key: "active",
		title: "Active webhooks",
		value: { target: 8 },
	},
	{
		icon: CircleCheck,
		key: "delivered",
		title: "Delivered",
		value: { target: 1248 },
	},
	{
		badge: { label: "Paused", variant: "outline" },
		icon: Pause,
		key: "paused",
		title: "Paused",
		value: { target: 2 },
	},
	{
		badge: { label: "Review", variant: "destructive" },
		icon: Activity,
		key: "attention",
		title: "Needs attention",
		value: { target: 3 },
	},
];

const meta = {
	argTypes: {
		ariaLabel: { control: "text" },
		className: { control: "text" },
		items: { control: false },
	},
	args: { ariaLabel: "Webhook overview", items },
	component: StatsCard,
	parameters: { layout: "padded" },
	title: "UI/StatsCard",
} satisfies Meta<typeof StatsCard>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Playground: Story = {};
export const ThreeMetrics: Story = { args: { items: items.slice(0, 3) } };
export const FiveMetrics: Story = {
	args: {
		items: [...items, { icon: Activity, key: "queued", title: "Queued", value: { target: 12 } }],
	},
};
export const SevenMetrics: Story = {
	args: {
		items: [
			...items,
			{ icon: Activity, key: "queued", title: "Queued", value: { target: 12 } },
			{ icon: Activity, key: "retried", title: "Retried", value: { target: 7 } },
			{ icon: Activity, key: "failed", title: "Failed", value: { target: 1 } },
		],
	},
};
export const ExactValues: Story = {
	args: {
		ariaLabel: "Payment overview",
		items: [
			{
				icon: Wallet,
				key: "volume",
				title: "Exact volume",
				value: { display: "Rp9.007.199.254.740.993" },
			},
			{
				icon: Activity,
				key: "change",
				title: "Balance change",
				value: {
					format: (value) =>
						new Intl.NumberFormat("id-ID", {
							currency: "IDR",
							maximumFractionDigits: 0,
							style: "currency",
						}).format(value),
					target: -125_000,
				},
			},
		],
	},
};
export const MissingValues: Story = {
	args: {
		items: [
			{
				caption: "No measurement yet",
				icon: Activity,
				key: "pending",
				title: "Awaiting data",
				value: { target: null },
			},
		],
	},
};
export const LongContent: Story = {
	args: {
		items: [
			{
				badge: {
					label: "Pending reconciliation across all payment providers",
					variant: "outline",
				},
				caption: "A long caption that wraps without hiding useful information on narrow screens.",
				icon: Activity,
				key: "long",
				title: "Total deliveries across all active production environments",
				value: { display: "Rp9.007.199.254.740.993" },
			},
		],
	},
};
export const Empty: Story = { args: { items: [] } };
export const SixMetrics: Story = {
	args: {
		items: [
			...items,
			{ icon: Activity, key: "queued", title: "Queued", value: { target: 12 } },
			{
				icon: Activity,
				key: "retried",
				title: "Retried",
				value: { target: 7 },
			},
		],
	},
};

const UpdatingExample = (args: StatsCardProps) => {
	const [target, setTarget] = useState(1248);
	return (
		<div className="flex flex-col items-start gap-4">
			<StatsCard
				{...args}
				items={[
					{
						icon: Activity,
						key: "delivered",
						title: "Delivered",
						value: { target },
					},
				]}
			/>
			<button
				type="button"
				className="rounded-md border border-border px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-ring"
				onClick={() => setTarget((current) => (current === 1248 ? 748 : 1248))}
			>
				Toggle value
			</button>
		</div>
	);
};
export const UpdatingValues: Story = {
	argTypes: { items: { control: false } },
	render: (args) => <UpdatingExample {...args} />,
};

/** A card that mounts with no value shows the final value once data arrives, in both motion modes. */
export const MissingThenValue: Story = {
	argTypes: { items: { control: false } },
	render: function MissingThenValueExample(args) {
		const [target, setTarget] = useState<number | null>(null);
		return (
			<div className="flex flex-col gap-4">
				<StatsCard {...args} items={[{ icon: Wallet, key: "late", title: "Late value", value: { target } }]} />
				<button type="button" className="w-fit rounded-md border border-border px-3 py-1 text-sm" onClick={() => setTarget(1248)}>
					Load value
				</button>
			</div>
		);
	},
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole("button", { name: "Load value" }));
		// Visible (non sr-only) value: the plain span in reduced motion, the animated span otherwise.
		const visible = () =>
			[...(canvasElement.querySelector("dd > div")?.children ?? [])]
				.filter((node) => !node.classList.contains("sr-only"))
				.map((node) => node.textContent)
				.join("");
		await waitFor(() => expect(visible()).toBe("1.248"), { timeout: 2000 });
	},
};
