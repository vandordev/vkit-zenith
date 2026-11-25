import type { Meta, StoryObj } from "@storybook/react";
import { OrganizationNotFound } from "./not-found";

const meta = {
	title: "Web/Pages/Dashboard/Entry/OrganizationNotFound",
	component: OrganizationNotFound,
	parameters: { layout: "fullscreen" },
} satisfies Meta<typeof OrganizationNotFound>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
/** The way back keeps the chosen mode. */
export const TestMode: Story = { args: { search: { mode: "test" } } };
