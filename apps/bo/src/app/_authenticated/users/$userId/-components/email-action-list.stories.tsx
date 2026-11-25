import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../-fixtures";
import { paged } from "../../../organizations/$organizationId/-setup-fixtures";
import { emailActions, now } from "../-fixtures";
import { EmailActionList } from "./email-action-list";

const meta = {
	title: "BO/Pages/Users/EmailActions",
	component: EmailActionList,
	parameters: { layout: "padded" },
	args: { actions: emailActions, state: "ready", pagination: paged(emailActions.length), now, onPage: noop, onRetry: noop },
	argTypes: { actions: { control: false } },
} satisfies Meta<typeof EmailActionList>;
export default meta;
type Story = StoryObj<typeof meta>;
/** Pending, consumed, revoked and expired actions with their null dates shown as a dash. */
export const Populated: Story = {};
export const Pending: Story = { args: { actions: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { actions: [], state: "error", error: "Tindakan email tidak dapat dimuat.", pagination: undefined } };
export const Empty: Story = { args: { actions: [], pagination: paged(0) } };
export const UnknownPurpose: Story = {
	args: { actions: [{ ...emailActions[0]!, purpose: "future_purpose" as never }], pagination: paged(1) },
};
export const Paging: Story = {
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return <EmailActionList {...args} pagination={paged(4, page, 4, "9")} onPage={setPage} />;
	},
};
