import type { Meta, StoryObj } from "@storybook/react";
import { RoutePending } from "./route-states";

const meta = {
	title: "Web/Components/RoutePending",
	component: RoutePending,
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof RoutePending>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
