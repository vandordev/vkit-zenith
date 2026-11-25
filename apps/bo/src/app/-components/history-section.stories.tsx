import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../-fixtures";
import { history } from "../-fixtures/audit";
import { paged } from "../_authenticated/organizations/$organizationId/-setup-fixtures";
import { HistorySection } from "./history-section";

const meta = {
	title: "BO/Components/HistorySection",
	component: HistorySection,
	parameters: { layout: "fullscreen" },
	args: {
		description: "Perubahan pada akun pelanggan ini beserta pelaku dan alasannya.",
		q: "",
		entries: history,
		state: "ready",
		pagination: paged(history.length),
		onSearch: noop,
		onPage: noop,
		onRetry: noop,
	},
	argTypes: { entries: { control: false } },
} satisfies Meta<typeof HistorySection>;
export default meta;
type Story = StoryObj<typeof meta>;
/** Includes the unknown future subsystem key and the 3-line reason from the shared audit fixtures. */
export const Populated: Story = {};
export const Pending: Story = { args: { entries: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { entries: [], state: "error", error: "Riwayat tidak dapat dimuat.", pagination: undefined } };
export const Empty: Story = { args: { entries: [], pagination: paged(0) } };
export const NoMatches: Story = { args: { q: "tidak-ada", entries: [], pagination: paged(0) } };
export const Refreshing: Story = { args: { q: "provider", loading: true, pagination: paged(25, 2, 25, "60"), entries: history.slice(0, 3) } };
export const LaterPageOutOfRange: Story = { args: { entries: [], pagination: paged(0, 9, 25, "60") } };
/** Real local state: typing filters after a short pause, Kosongkan filter resets at once, no search button exists. */
export const Interactive: Story = {
	render: function Preview(args) {
		const [q, setQ] = useState("");
		const entries = history.filter((entry) => !q || `${entry.action} ${entry.actor.name ?? ""} ${entry.actor.system ?? ""}`.toLowerCase().includes(q.toLowerCase()));
		return <HistorySection {...args} q={q} entries={entries} pagination={paged(entries.length)} onSearch={setQ} />;
	},
};
