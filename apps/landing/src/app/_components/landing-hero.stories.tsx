import type { Meta, StoryObj } from "@storybook/react";
import { LandingHero } from "./landing-hero";
import { withLandingFrame } from "./story-frame";

const meta = {
	title: "Landing/Components/Hero",
	component: LandingHero,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof LandingHero>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
