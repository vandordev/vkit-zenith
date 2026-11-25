import type { Meta, StoryObj } from "@storybook/react";
import { FeatureBento } from "./feature-bento";
import { withLandingFrame } from "./story-frame";

const meta = {
	title: "Landing/Components/FeatureBento",
	component: FeatureBento,
	decorators: [withLandingFrame],
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof FeatureBento>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
