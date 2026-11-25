import type { Meta, StoryObj } from "@storybook/react";
import { OrgOverview } from "./org-overview";
import { longOrganization, organization } from "./-fixtures";

const noop = () => {};
const meta = {
	title: "Web/Pages/Dashboard/Overview",
	component: OrgOverview,
	parameters: { layout: "fullscreen", routerInitialPath: "/w/kopi" },
	args: { organization: organization("owner"), memberCount: 4, membersState: "ready", onRetryMembers: noop },
	argTypes: { organization: { control: false } },
} satisfies Meta<typeof OrgOverview>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Owner: Story = {};
/** Members cannot manage API keys, so that shortcut is absent. */
export const Member: Story = { args: { organization: organization("member") } };
export const Suspended: Story = { args: { organization: organization("owner", { status: "suspended" }) } };
export const LongName: Story = { args: { organization: longOrganization } };
export const Pending: Story = { args: { memberCount: undefined, membersState: "pending" } };
export const Error: Story = {
	args: { memberCount: undefined, membersState: "error", membersError: "Layanan sedang tidak tersedia. Coba lagi sebentar lagi." },
};
