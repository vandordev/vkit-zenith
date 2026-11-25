import type { Meta, StoryObj } from "@storybook/react";
import { RecentActivity } from "./recent-activity";
import { event, summary, missing, noop, customerEvent } from "../activity/-fixtures";
const recentProps = {
	data: { items: [event, summary, missing] },
	onRetry: noop,
	renderLink: (row: typeof event) => <a href={`#${row.id}`}>Periksa aktivitas</a>,
	renderAll: () => <a href="#activity">Lihat log aktivitas</a>,
};
const meta = {
	title: "BO/Pages/Home/RecentActivity",
	component: RecentActivity,
	args: recentProps,
	parameters: { layout: "padded" },
} satisfies Meta<typeof RecentActivity>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const CustomerEvent: Story = { args: { data: { items: [customerEvent, event, summary] } } };
export const Pending: Story = { args: { data: undefined, loading: true } };
export const Error: Story = { args: { data: undefined, error: "Aktivitas tidak tersedia." } };
export const Empty: Story = { args: { data: { items: [] } } };
export const RetainedError: Story = { args: { error: "Pembaruan aktivitas gagal. Data sebelumnya masih ditampilkan." } };
