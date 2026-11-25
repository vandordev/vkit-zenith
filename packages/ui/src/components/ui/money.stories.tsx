import type { Meta, StoryObj } from "@storybook/react";

import { Money } from "./money";

const meta = {
	args: { amount: "125000000", currency: "IDR" },
	argTypes: { sign: { control: "inline-radio", options: ["auto", "negative"] } },
	component: Money,
	parameters: { layout: "centered" },
	title: "UI/Money",
} satisfies Meta<typeof Money>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Idr: Story = {};
export const IdrZero: Story = { args: { amount: "0" } };
export const UsdTwoDecimals: Story = { args: { amount: "1999", currency: "USD" } };
export const NegativeRefund: Story = { args: { amount: "25000000", sign: "negative" } };
export const NegativeInput: Story = { args: { amount: "-25000000" } };
export const BeyondSafeInteger: Story = {
	args: { amount: "9007199254740993" },
	parameters: { docs: { description: { story: "Exact beyond 2^53: must read Rp9.007.199.254.740.993." } } },
};
export const InvalidAmount: Story = { args: { amount: "abc" } };
export const InvalidCurrency: Story = { args: { currency: "XXX1" } };
export const TabularFigures: Story = {
	render: () => (
		<div className="flex flex-col items-end gap-1 text-lg">
			<Money amount="111111111" currency="IDR" />
			<Money amount="100000000" currency="IDR" />
			<Money amount="888888888" currency="IDR" />
		</div>
	),
	parameters: { docs: { description: { story: "Digits must align in columns." } } },
};
