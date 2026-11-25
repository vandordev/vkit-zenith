import type { Meta, StoryObj } from "@storybook/react";
import { Link } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { FilterBar } from "@repo/ui/filter-bar";
import { filterLabels } from "./filter-labels";
import { ListPage, type ListPageTab } from "./list-page";
import { StatusBadge } from "./status-badge";

const noop = () => {};
type Row = { id: string; name: string; status: "active" | "expired" | "revoked" };
const rows: Row[] = [
	{ id: "key_0001", name: "Server produksi", status: "active" },
	{ id: "key_0002", name: "Staging", status: "expired" },
	{ id: "key_0003_dengan_referensi_yang_sangat_panjang_sekali_untuk_menguji_overflow", name: "Integrasi lama", status: "revoked" },
];
// The story renders the generic page for this row type.
const RowsPage = ListPage<Row>;
const tabs: ListPageTab[] = [
	{ key: "all", label: "Semua", to: ".", params: {}, active: true },
	{ key: "active", label: "Aktif", to: ".", params: {}, active: false },
];

const meta = {
	title: "Web/Components/Dashboard/ListPage",
	component: RowsPage,
	parameters: { layout: "fullscreen" },
	args: {
		title: "Daftar",
		description: "Contoh halaman daftar generik.",
		primaryAction: <Button type="button">Aksi utama</Button>,
		tabs,
		filters: (
			<FilterBar
				search={{ value: "", onChange: noop, placeholder: "Cari…" }}
				filters={[{ id: "status", label: "Status", options: [{ value: "active", label: "Aktif" }], value: "", onChange: noop }]}
				labels={filterLabels}
			/>
		),
		items: rows,
		getKey: (row) => row.id,
		getLink: () => <Link to="." />,
		columns: [
			{ id: "id", header: "ID", cell: (row) => row.id },
			{ id: "name", header: "Nama", cell: (row) => row.name },
			{ id: "status", header: "Status", cell: (row) => <StatusBadge resource="apiKey" status={row.status} /> },
		],
		pagination: { page: 1, size: 20, total: "3" },
		noun: "data",
		state: "ready",
		emptyTitle: "Belum ada data",
		emptyDescription: "Data pertama akan muncul di sini.",
		onPage: noop,
		onRetry: noop,
	},
	argTypes: {
		primaryAction: { control: false },
		filters: { control: false },
		items: { control: false },
		columns: { control: false },
		getKey: { control: false },
		getLink: { control: false },
		tabs: { control: false },
		pagination: { control: false },
	},
} satisfies Meta<typeof RowsPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {};
export const Paged: Story = { args: { pagination: { page: 2, size: 3, total: "11" } } };
export const Empty: Story = { args: { items: [], pagination: { page: 1, size: 20, total: "0" } } };
/** Page 3 of a list that now has 5 records: empty copy for the page, Previous still reachable. */
export const OutOfRange: Story = { args: { items: [], pagination: { page: 3, size: 3, total: "5" } } };
/** The last record vanished while page 3 was open: total is 0 but the way back to page 1 stays. */
export const ZeroTotalOutOfRange: Story = { args: { items: [], pagination: { page: 3, size: 3, total: "0" } } };
export const NoMatches: Story = { args: { items: [], filtered: true, pagination: { page: 1, size: 20, total: "0" } } };
export const Pending: Story = { args: { state: "pending", items: [] } };
export const Error: Story = {
	args: { state: "error", items: [], errorMessage: "Layanan sedang tidak tersedia. Coba lagi sebentar lagi." },
};
/** Shows a dummy write result instead of refetched data. */
export const Simulated: Story = { args: { simulated: true } };
