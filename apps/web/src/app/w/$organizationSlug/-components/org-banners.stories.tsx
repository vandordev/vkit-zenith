import type { Meta, StoryObj } from "@storybook/react";
import { OrganizationBanners } from "./org-banners";

const meta = {
	title: "Web/Components/Dashboard/OrganizationBanners",
	component: OrganizationBanners,
	args: { suspended: false },
} satisfies Meta<typeof OrganizationBanners>;
export default meta;
type Story = StoryObj<typeof meta>;

export const None: Story = {};
export const Suspended: Story = { args: { suspended: true } };
