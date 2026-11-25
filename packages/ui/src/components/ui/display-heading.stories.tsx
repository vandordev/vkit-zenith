import type { Meta, StoryObj } from "@storybook/react";

import { DisplayHeading } from "./display-heading";

const meta = {
	args: { children: "Masuk untuk bekerja" },
	argTypes: { as: { control: "inline-radio", options: ["h1", "h2"] } },
	component: DisplayHeading,
	parameters: { layout: "padded" },
	title: "UI/DisplayHeading",
} satisfies Meta<typeof DisplayHeading>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
export const SecondLevel: Story = { args: { as: "h2" } };
export const LongContent: Story = {
	args: { children: "Judul yang sangat panjang untuk memastikan tidak ada luapan pada layar sempit termasuk kata-kata-yang-sangat-panjang-tanpa-spasi" },
	decorators: [(Story) => <div className="max-w-md"><Story /></div>],
};
