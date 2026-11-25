import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { FilterBar } from "./filter-bar";

const statuses = [
	{ value: "succeeded", label: "Berhasil" },
	{ value: "pending", label: "Menunggu" },
	{ value: "failed", label: "Gagal" },
];
const methods = [
	{ value: "bank_transfer", label: "Transfer bank" },
	{ value: "qris", label: "QRIS" },
	{ value: "card", label: "Kartu" },
];
const sorts = [
	{ value: "created:desc", label: "Terbaru" },
	{ value: "created:asc", label: "Terlama" },
	{ value: "amount:desc", label: "Jumlah terbesar" },
];
const labels = {
	filter: "Filter",
	filters: "Filter dan urutan",
	clear: "Kosongkan",
	sort: "Urutkan",
	all: "Semua",
	remove: (field: string) => `Hapus filter ${field.toLowerCase()}`,
	change: (field: string) => `Ubah filter ${field.toLowerCase()}`,
};

const Preview = ({ initial = {}, withSort = true }: { initial?: { q?: string; status?: string; method?: string }; withSort?: boolean }) => {
	const [q, setQ] = useState(initial.q ?? "");
	const [status, setStatus] = useState(initial.status ?? "");
	const [method, setMethod] = useState(initial.method ?? "");
	const [sort, setSort] = useState("created:desc");
	return (
		<div className="flex flex-col gap-4">
			<FilterBar
				search={{ value: q, onChange: setQ, placeholder: "Cari pembayaran…" }}
				filters={[
					{ id: "status", label: "Status", options: statuses, value: status, onChange: setStatus },
					{ id: "method", label: "Metode", options: methods, value: method, onChange: setMethod },
				]}
				sort={withSort ? { value: sort, options: sorts, onChange: setSort } : undefined}
				onClear={() => {
					setQ("");
					setStatus("");
					setMethod("");
				}}
				labels={labels}
			/>
			<p role="status" className="text-xs text-muted-foreground">
				q="{q}" status="{status}" method="{method}" sort="{sort}"
			</p>
		</div>
	);
};

const meta = {
	title: "UI/FilterBar",
	component: FilterBar,
	parameters: { layout: "padded" },
	render: () => <Preview />,
} satisfies Meta<typeof FilterBar>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
/** Active filters become chips: the label part changes the value, × removes it. */
export const Active: Story = { render: () => <Preview initial={{ q: "budi", status: "succeeded", method: "qris" }} /> };
/** Without a supported sort the bar shows no sort control. */
export const WithoutSort: Story = { render: () => <Preview withSort={false} /> };
/** Under 36rem of its own width, filters and sort move into a drawer; chips stay. */
export const Narrow: Story = {
	render: () => <Preview initial={{ status: "pending" }} />,
	decorators: [(Story) => <div className="max-w-sm">{Story()}</div>],
};
