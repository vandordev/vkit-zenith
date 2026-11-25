import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../../-fixtures";
import { paged } from "../../../../organizations/$organizationId/-setup-fixtures";
import { now, userSessions } from "../../-fixtures";
import { UserSessions } from "./user-sessions";

const meta = {
	title: "BO/Pages/Users/Sessions",
	component: UserSessions,
	parameters: { layout: "fullscreen" },
	args: { sessions: userSessions, state: "ready", pagination: paged(userSessions.length), now, onPage: noop, onRetry: noop },
	argTypes: { sessions: { control: false } },
} satisfies Meta<typeof UserSessions>;
export default meta;
type Story = StoryObj<typeof meta>;
/** Active, expired and revoked sessions side by side. */
export const Populated: Story = {};
export const Pending: Story = { args: { sessions: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { sessions: [], state: "error", error: "Sesi tidak dapat dimuat.", pagination: undefined } };
export const Empty: Story = { args: { sessions: [], pagination: paged(0) } };
export const LaterPageOutOfRange: Story = { args: { sessions: [], pagination: paged(0, 9, 25, "60") } };
export const Paging: Story = {
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return <UserSessions {...args} pagination={paged(25, page, 25, "60")} onPage={setPage} />;
	},
};
