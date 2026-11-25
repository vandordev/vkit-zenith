import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../../-fixtures";
import { history } from "../../../../../-fixtures/audit";
import { paged } from "../../-setup-fixtures";
import { OrganizationActivity } from "./organization-activity";

const meta = {
	title: "BO/Pages/Organizations/Activity",
	component: OrganizationActivity,
	parameters: { layout: "fullscreen" },
	args: { entries: history, state: "ready", pagination: paged(history.length), onPage: noop, onRetry: noop },
	argTypes: { entries: { control: false } },
} satisfies Meta<typeof OrganizationActivity>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Empty: Story = { args: { entries: [], state: "ready", pagination: paged(0) } };
export const Pending: Story = { args: { entries: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { entries: [], state: "error", error: "Data dummy gagal dimuat", pagination: undefined } };
export const LaterPage: Story = { args: { entries: history.slice(0, 3), pagination: paged(3, 2, 25, "60") } };
export const LaterPageOutOfRange: Story = { args: { entries: [], pagination: paged(0, 9, 25, "60") } };
export const Paging: Story = {
	args: { pagination: paged(25, 1, 25, "60") },
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return <OrganizationActivity {...args} pagination={paged(25, page, 25, "60")} onPage={setPage} />;
	},
};
