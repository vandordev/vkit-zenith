import type { Meta, StoryObj } from "@storybook/react";
import { JobCollection } from "./job-collection";
import { collectionArgs, jobs, pageOf } from "../../../-global-fixtures";

const live = jobs;
const meta = {
	title: "BO/Pages/Operations/JobCollection",
	component: JobCollection,
	parameters: { layout: "fullscreen", routerInitialPath: "/operations/jobs" },
	args: { ...collectionArgs, items: live, pagination: pageOf(live.length), status: "", stateGroup: "" },
	argTypes: { items: { control: false }, pagination: { control: false } },
} satisfies Meta<typeof JobCollection>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {};
export const Empty: Story = { args: { items: [], pagination: pageOf(0) } };
export const NoMatches: Story = { args: { items: [], pagination: pageOf(0), q: "tidak-ada" } };
export const Pending: Story = { args: { items: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { items: [], state: "error", error: "Data tidak dapat dimuat (503).", pagination: undefined } };
export const LaterPageEmpty: Story = { args: { items: [], pagination: pageOf(0, 3, "60") } };
export const FailedGroup: Story = { args: { stateGroup: "failed", items: live.filter((job) => job.state === "retryable" || job.state === "discarded"), pagination: pageOf(2) }, parameters: { routerInitialPath: "/operations/jobs?stateGroup=failed" } };
