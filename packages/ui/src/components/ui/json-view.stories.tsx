import type { Meta, StoryObj } from "@storybook/react";

import { JsonView } from "./json-view";

const meta = {
	args: { label: "Payload webhook", value: { id: "evt_01", type: "payment.succeeded", amount: "125000000", currency: "IDR", metadata: null } },
	component: JsonView,
	decorators: [(Story) => <div className="w-[28rem] max-w-full"><Story /></div>],
	parameters: { layout: "centered" },
	title: "UI/JsonView",
} satisfies Meta<typeof JsonView>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Short: Story = {};
export const LongCollapsed: Story = {
	args: { value: { items: Array.from({ length: 30 }, (_, index) => ({ index, name: `Item ${index + 1}`, amount: String(index * 1000) })) } },
};
export const WideLine: Story = {
	args: { value: { description: "x".repeat(200) } },
};
export const Primitive: Story = { args: { value: null } };
