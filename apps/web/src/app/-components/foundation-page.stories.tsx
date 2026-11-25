import type { Meta, StoryObj } from "@storybook/react";
import { FoundationPage } from "./foundation-page";

const meta = {
	title: "Web/Pages/Foundation",
	component: FoundationPage,
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof FoundationPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Narrow: Story = {
	decorators: [(Story) => <div className="mx-auto max-w-[22.5rem]"><Story /></div>],
};
