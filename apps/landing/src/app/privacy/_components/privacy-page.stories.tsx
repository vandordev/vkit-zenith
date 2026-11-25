import type { Meta, StoryObj } from "@storybook/react";
import { withLandingFrame } from "../../_components/story-frame";
import { PrivacyPage } from "./privacy-page";

const meta = {
	title: "Landing/Pages/Privacy",
	component: PrivacyPage,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof PrivacyPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
