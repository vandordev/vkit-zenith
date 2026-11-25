import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../../-fixtures";
import { paged } from "../../../../organizations/$organizationId/-setup-fixtures";
import { staffSessions } from "../../../-fixtures";
import { StaffSessions } from "./staff-sessions";

const now = Date.parse("2026-11-09T12:00:00Z");
const meta = {
	title: "BO/Pages/Staff/Sessions",
	component: StaffSessions,
	parameters: { layout: "fullscreen" },
	args: { sessions: staffSessions, state: "ready", pagination: paged(staffSessions.length), now, onPage: noop, onRetry: noop },
	argTypes: { sessions: { control: false } },
} satisfies Meta<typeof StaffSessions>;
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
		return <StaffSessions {...args} pagination={paged(25, page, 25, "60")} onPage={setPage} />;
	},
};
