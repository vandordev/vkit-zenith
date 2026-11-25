import type { Meta, StoryObj } from "@storybook/react";
import { LandingNav } from "./landing-nav";
import { withLandingFrame } from "./story-frame";

const meta = {
	title: "Landing/Components/Nav",
	component: LandingNav,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof LandingNav>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
