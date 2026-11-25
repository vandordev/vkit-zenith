import type { Meta, StoryObj } from "@storybook/react";
import { AuditTimeline } from "./audit-timeline";
import { history } from "../-fixtures/audit";
import { noop } from "../-fixtures";

const meta = {
	title: "BO/Components/AuditTimeline",
	component: AuditTimeline,
	args: { entries: history, state: "ready", onRetry: noop },
	argTypes: { entries: { control: false }, onRetry: { control: false } },
	decorators: [
		(Story) => (
			<div className="mx-auto max-w-3xl p-4">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof AuditTimeline>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Pending: Story = { args: { entries: [], state: "pending" } };
export const Error: Story = { args: { entries: [], state: "error" } };
export const Empty: Story = { args: { entries: [], state: "empty" } };
export const SingleEntry: Story = { args: { entries: history.slice(0, 1) } };
