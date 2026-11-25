import type { Meta, StoryObj } from "@storybook/react";
import { BOError } from "@bo/lib/api/client";
import { JobDetail } from "./job-detail";
import { jobs, delayed } from "../../../-global-fixtures";
import { counted, detailArgs } from "../../-components/ops-story-helpers";

const find = (state: string) => jobs.find((j) => j.state === state)!;
const retryable = find("retryable");
const args = (item = retryable) => ({ ...detailArgs(item, "job"), onRetryJob: async () => ({ ...item, state: "available" }), onCancel: async () => ({ ...item, state: "cancelled" }) });
const meta = {
	title: "BO/Pages/Operations/JobDetail",
	component: JobDetail,
	parameters: { layout: "fullscreen", routerInitialPath: `/operations/jobs/${retryable.id}` },
	args: args(),
	argTypes: { item: { control: false }, history: { control: false } },
} satisfies Meta<typeof JobDetail>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Retryable offers both Ulangi and Batalkan; each other state shows only what the backend allows. */
export const Retryable: Story = {};
export const Discarded: Story = { args: args(find("discarded")) };
export const Available: Story = { args: args(find("available")) };
export const Scheduled: Story = { args: args(find("scheduled")) };
export const Running: Story = { args: args(find("running")) };
export const Completed: Story = { args: args(find("completed")) };
export const Cancelled: Story = { args: args(find("cancelled")) };
export const LongKind: Story = { args: args(jobs.find((j) => j.kind.length > 50)!) };
export const HistoryEmpty: Story = { args: { history: [] } };
export const Pending: Story = { args: { state: "pending", item: undefined } };
export const Empty: Story = { args: { state: "empty", item: undefined } };
export const Error: Story = { args: { state: "error", error: "Detail job tidak dapat dimuat", item: undefined } };
/** Retry answers after 600 ms: the job turns available, Ulangi disappears, the dialog result stays until Tutup. */
export const RetryLifecycle: Story = {
	render: (a) => {
		const Preview = counted(JobDetail, a, (count) => ({ onRetryJob: async () => { count(); return delayed({ ...retryable, state: "available" }); } }));
		return <Preview />;
	},
};
export const CancelLifecycle: Story = {
	render: (a) => {
		const Preview = counted(JobDetail, a, (count) => ({ onCancel: async () => { count(); return delayed({ ...retryable, state: "cancelled" }); } }));
		return <Preview />;
	},
};
export const Conflict409: Story = {
	render: (a) => {
		const Preview = counted(JobDetail, a, (count) => ({ onRetryJob: async () => { count(); throw new BOError(409, "conflict", "Job sudah berubah. Muat ulang halaman."); } }));
		return <Preview />;
	},
};
export const Unprocessable422: Story = {
	render: (a) => {
		const Preview = counted(JobDetail, a, (count) => ({ onCancel: async () => { count(); await delayed(0, 1500); throw new BOError(422, "validation_failed", "Alasan tidak valid."); } }));
		return <Preview />;
	},
};

