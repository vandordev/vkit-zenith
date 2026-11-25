import type { Meta, StoryObj } from "@storybook/react";
import { expect, userEvent, waitFor, within } from "storybook/test";

import { CopyableId } from "./copyable-id";

const meta = {
	args: { label: "ID organisasi", value: "org_01HZX9" },
	argTypes: { truncate: { control: "inline-radio", options: ["none", "middle"] } },
	component: CopyableId,
	parameters: { layout: "centered" },
	title: "UI/CopyableId",
} satisfies Meta<typeof CopyableId>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Short: Story = {};
export const LongMiddleTruncated: Story = {
	args: { label: "ID pembayaran", truncate: "middle", value: "pay_01HZX9K3M7Q2V8N4R6T1W5Y0AB" },
	decorators: [(Story) => <div className="w-48"><Story /></div>],
};
export const LongNoTruncate: Story = {
	args: { label: "ID pembayaran", value: "pay_01HZX9K3M7Q2V8N4R6T1W5Y0AB01HZX9K3M7Q2V8N4R6T1W5Y0AB" },
	decorators: [(Story) => <div className="w-48"><Story /></div>],
};
export const CustomDisplay: Story = { args: { display: "•••• 1234", label: "nomor rekening", value: "1234" } };
export const Copied: Story = {
	play: async ({ canvasElement }) => {
		const button = within(canvasElement).getByRole("button");
		// setup() installs user-event's clipboard stub, so the story does not depend on browser clipboard permission.
		await userEvent.setup().click(button);
		await waitFor(() => expect(button).toHaveAccessibleName("ID organisasi disalin"));
	},
};
