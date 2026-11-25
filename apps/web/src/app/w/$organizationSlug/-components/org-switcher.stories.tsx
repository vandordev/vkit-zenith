import type { Meta, StoryObj } from "@storybook/react";
import { OrgSwitcher } from "./org-switcher";
import { longOrganization, organization, organizations } from "./-fixtures";

const meta = {
	title: "Web/Components/Dashboard/OrgSwitcher",
	component: OrgSwitcher,
	args: { organization: organization(), organizations: [organization()] },
	argTypes: { organization: { control: false }, organizations: { control: false } },
	decorators: [
		(Story) => (
			<div className="w-60 border border-border bg-sidebar">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof OrgSwitcher>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Single: Story = {};
export const Several: Story = { args: { organizations } };
export const LongName: Story = { args: { organization: longOrganization, organizations: [longOrganization, ...organizations.slice(1)] } };
export const Collapsed: Story = {
	args: { organizations, collapsed: true },
	decorators: [
		(Story) => (
			<div className="w-12 border border-border bg-sidebar">
				<Story />
			</div>
		),
	],
};
