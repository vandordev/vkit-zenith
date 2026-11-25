import type { Meta, StoryObj } from "@storybook/react";
import { LandingFooter } from "./landing-footer";
import { withLandingFrame } from "./story-frame";

const meta = {
	title: "Landing/Components/Footer",
	component: LandingFooter,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof LandingFooter>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
