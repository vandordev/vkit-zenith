import type { Meta, StoryObj } from "@storybook/react";
import { HomeOverview, counterTarget } from "./home-counters";
import { RecentActivity } from "./recent-activity";
import { counters } from "../-global-fixtures";
import { event, summary, noop } from "../activity/-fixtures";

const activity = (
	<RecentActivity data={{ items: [event, summary] }} onRetry={noop} renderLink={(row) => <a href={`#${row.id}`}>Periksa aktivitas</a>} renderAll={() => <a href="#activity">Lihat log aktivitas</a>} />
);
const meta = {
	title: "BO/Pages/Home/Counters",
	component: HomeOverview,
	parameters: { layout: "fullscreen", routerInitialPath: "/" },
	args: { counters: counters("3"), state: "ready", onRetry: noop, children: activity },
	argTypes: { counters: { control: false }, children: { control: false } },
} satisfies Meta<typeof HomeOverview>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {};
/** Admin: counters only, no activity (the activity log is superadmin-only). */
export const AdminWithoutActivity: Story = { args: { children: null } };
export const AllZero: Story = { args: { counters: counters("0") } };
export const LargeCounts: Story = { args: { counters: counters("1234567890123") } };
export const Pending: Story = { args: { counters: undefined, state: "pending" } };
export const Error: Story = { args: { counters: undefined, state: "error", error: "Penghitung tidak dapat dimuat (503)." } };
/** A destination the validator would change (unknown status) gets no link instead of a mismatched list. */
export const UnlinkableCounter: Story = {
	args: { counters: { failedJobs: { count: "1", href: "https://evil.example/operations/jobs" } } },
};
export const TargetsMatchCounters: Story = {
	render: () => (
		<pre className="p-4 text-xs">
			{Object.values(counters("1")).map((counter) => JSON.stringify(counterTarget(counter.href))).join("\n")}
		</pre>
	),
};
