import type { Meta, StoryObj } from "@storybook/react";
import { GlobalPending } from "./global-pending";

const meta = {
	title: "Web/Components/GlobalPending",
	component: GlobalPending,
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof GlobalPending>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
