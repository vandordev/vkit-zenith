import type { Meta, StoryObj } from "@storybook/react";
import { HistorySection } from "../../../../-components/history-section";
import { noop } from "../../../../-fixtures";
import { staffHistory } from "../../-fixtures";

const meta = {
	title: "BO/Pages/StaffActivity",
	component: HistorySection,
	parameters: { layout: "fullscreen" },
	args: {
		description: "Siapa atau apa yang mengubah akun staf ini, beserta alasannya.",
		q: "",
		entries: staffHistory,
		state: "ready",
		pagination: { page: 1, pageSize: 25, total: String(staffHistory.length) },
		onSearch: noop,
		onPage: noop,
		onRetry: noop,
	},
} satisfies Meta<typeof HistorySection>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {};
export const Loading: Story = { args: { entries: [], state: "pending", pagination: undefined } };
export const Empty: Story = { args: { entries: [], state: "ready", pagination: { page: 1, pageSize: 25, total: "0" } } };
export const Error: Story = { args: { entries: [], state: "error", error: "Riwayat aktivitas tidak dapat dimuat.", pagination: undefined } };
