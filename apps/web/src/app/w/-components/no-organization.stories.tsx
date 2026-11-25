import type { Meta, StoryObj } from "@storybook/react";
import { noop, useSubmitCycle } from "../../-story-helpers";
import { NoOrganization } from "./no-organization";

const meta = {
	title: "Web/Pages/Dashboard/Entry/NoOrganization",
	component: NoOrganization,
	parameters: { layout: "fullscreen" },
	args: { displayName: "Rani Wulandari", logout: { kind: "idle" }, onLogout: noop },
	argTypes: { logout: { control: false } },
} satisfies Meta<typeof NoOrganization>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const LoggingOut: Story = { args: { logout: { kind: "pending" } } };
export const LogoutFailed: Story = { args: { logout: { kind: "unavailable" } } };
export const Interactive: Story = {
	render: (args) => {
		const [logout, submit] = useSubmitCycle({ kind: "unavailable" });
		return <NoOrganization {...args} logout={logout} onLogout={submit} />;
	},
};
