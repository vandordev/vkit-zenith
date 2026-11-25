import type { Meta, StoryObj } from "@storybook/react";
import { LandingPage } from "./landing-page";
import { withLandingFrame } from "./story-frame";

const meta = {
	title: "Landing/Pages/OnePage",
	component: LandingPage,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof LandingPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
