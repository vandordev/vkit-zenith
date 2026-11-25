import type { Meta, StoryObj } from "@storybook/react";
import { Link, useLocation, useRouter } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { ListFilters } from "./global-list";
import { statusOptions } from "./filter-options";
import { noop } from "../../-fixtures";

const base = {
	searchLabel: "Cari job",
	q: "",
	selects: [{ key: "status", label: "Status", value: "", options: statusOptions(["retryable", "completed", "discarded"], "job") }],
	onFilter: noop,
};
const interactiveStatuses = ["", "retryable", "completed", "discarded"] as const;
type InteractiveStatus = (typeof interactiveStatuses)[number];
function parseInteractiveStatus(value: string | undefined): InteractiveStatus {
	if (value === "retryable" || value === "completed" || value === "discarded") return value;
	return "";
}
const meta = {
	title: "BO/Components/GlobalListFilters",
	component: ListFilters,
	parameters: { layout: "padded", routerInitialPath: "/operations/jobs" },
	args: base,
	argTypes: { selects: { control: false } },
} satisfies Meta<typeof ListFilters>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Filtered: Story = {
	args: { q: "email", selects: [{ ...base.selects[0]!, value: "discarded" }], organizationId: "00000000-0000-4000-8000-000000000001", onClearOrganization: noop },
};
/** The real filter composition writes validated search state into Storybook's memory history. */
export const Interactive: Story = {
	parameters: { routerInitialPath: "/operations/jobs?q=email&status=completed&page=3" },
	render: function Preview(args) {
		const location = useLocation();
		const router = useRouter();
		const search = location.search as { q?: string; status?: string; page?: number; organizationId?: string };
		const q = search.q ?? "";
		const status = parseInteractiveStatus(search.status);
		const page = search.page ?? 1;
		return (
			<div className="flex flex-col gap-4">
				<ListFilters
					{...args}
					q={q}
					selects={[{ ...base.selects[0]!, value: status }]}
					navigationKey={JSON.stringify([search.page, search.organizationId, search.status])}
					onFilter={(next, values) => void router.navigate({ to: ".", search: (previous) => ({ ...previous, q: next, status: parseInteractiveStatus(values.status), page: 1 }) })}
				/>
				<div className="flex flex-wrap gap-2">
					<Link to="." search={(previous) => ({ ...previous, page: 2 })}>Halaman 2</Link>
					<Button type="button" variant="outline" onClick={() => router.history.back()}>Kembali di riwayat pratinjau</Button>
					<Button type="button" variant="outline" onClick={() => router.history.forward()}>Maju di riwayat pratinjau</Button>
				</div>
				<p role="status" className="text-sm text-muted-foreground">
					Diterapkan: q="{q}" status="{status}" page="{page}" organisasi="{search.organizationId}"
				</p>
			</div>
		);
	},
};
