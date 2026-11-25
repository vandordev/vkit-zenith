import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { Badge } from "./badge";
import { BoringAvatar } from "./boring-avatar";
import { Button } from "./button";
import { Collection, type CollectionColumn, type CollectionSort } from "./collection";

type Payment = {
	id: string;
	name: string;
	email?: string;
	method: string;
	amount: number;
	status: [string, "success-light" | "warning-light" | "destructive-light"];
	date: string;
};
const payments: Payment[] = [
	{
		id: "pay_0001",
		name: "Budi Santoso",
		email: "budi@example.test",
		method: "Transfer bank",
		amount: 1500000,
		status: ["Berhasil", "success-light"],
		date: "2026-11-18",
	},
	{
		id: "pay_0002",
		name: "Sari Wulandari",
		email: "sari@example.test",
		method: "QRIS",
		amount: 250000,
		status: ["Menunggu", "warning-light"],
		date: "2026-11-18",
	},
	{
		id: "pay_0003",
		name: "Pembeli tanpa nama",
		method: "Kartu",
		amount: 89000,
		status: ["Gagal", "destructive-light"],
		date: "2026-11-17",
	},
	{
		id: "pay_0004",
		name: "PT Nusantara Digital Kreatif Indonesia Sejahtera",
		email: "finance@nusantara.example.test",
		method: "Virtual account",
		amount: 12750000,
		status: ["Berhasil", "success-light"],
		date: "2026-11-16",
	},
];
const rupiah = (value: number) => `Rp${value.toLocaleString("id-ID")}`;
const day = (value: string) => new Date(value).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
const paymentColumns: CollectionColumn<Payment>[] = [
	{
		id: "buyer",
		header: "Pembeli",
		cell: (payment) => (
			<span className="flex flex-col gap-0.5">
				<span>{payment.name}</span>
				{payment.email && <span className="text-xs font-normal text-muted-foreground">{payment.email}</span>}
			</span>
		),
	},
	{ id: "method", header: "Metode", cell: (payment) => payment.method },
	{ id: "amount", header: "Jumlah", align: "end", sortable: true, cell: (payment) => rupiah(payment.amount) },
	{ id: "status", header: "Status", cell: (payment) => <Badge variant={payment.status[1]}>{payment.status[0]}</Badge> },
	{ id: "created", header: "Dibuat", sortable: true, cell: (payment) => day(payment.date) },
];

const SortablePayments = () => {
	const [sort, setSort] = useState<CollectionSort>({ id: "created", direction: "desc" });
	const sign = sort.direction === "asc" ? 1 : -1;
	const sorted = [...payments].sort((a, b) => sign * (sort.id === "amount" ? a.amount - b.amount : a.date.localeCompare(b.date)));
	return (
		<Collection
			label="Pembayaran"
			items={sorted}
			columns={paymentColumns}
			getKey={(payment) => payment.id}
			getLink={(payment) => <a href={`#${payment.id}`} />}
			sort={sort}
			onSortChange={setSort}
			labels={{ sortAscending: "urutkan naik", sortDescending: "urutkan turun" }}
		/>
	);
};

type Customer = { id: string; name: string; email: string; phone?: string; externalId?: string };
const customers: Customer[] = [
	{ id: "cus_1", name: "Budi Santoso", email: "budi@example.test", phone: "+62 812-3456-7890", externalId: "CRM-1042" },
	{ id: "cus_2", name: "Sari Wulandari", email: "sari@example.test", phone: "+62 813-2222-1111" },
	{ id: "cus_3", name: "Rizky Pratama", email: "rizky.pratama@example.test", externalId: "CRM-2210" },
	{
		id: "cus_4",
		name: "Dewi Lestari Anggraini Kusumawardhani",
		email: "dewi@example.test",
		phone: "+62 811-0000-1234",
		externalId: "CRM-3301",
	},
	{ id: "cus_5", name: "Andi Wijaya", email: "andi@example.test" },
];
const customerColumns: CollectionColumn<Customer>[] = [
	{ id: "name", header: "Pelanggan", cell: (customer) => customer.name },
	{ id: "avatar", header: "Avatar", slot: "media", cell: (customer) => <BoringAvatar name={customer.id} variant="beam" size={40} /> },
	{ id: "email", header: "Email", cell: (customer) => customer.email },
	{ id: "phone", header: "Telepon", cell: (customer) => customer.phone ?? "—" },
	{ id: "external", header: "ID eksternal", inlineLabel: true, cell: (customer) => customer.externalId ?? "—" },
];

type Key = { id: string; name: string; prefix: string; created: string; used?: string; active: boolean };
const keys: Key[] = [
	{ id: "key_1", name: "Server produksi", prefix: "zk_a1b2", created: "2026-11-12", used: "2026-11-20", active: true },
	{ id: "key_2", name: "Staging", prefix: "zk_c3d4", created: "2026-11-03", active: true },
	{ id: "key_3", name: "Integrasi lama", prefix: "zk_e5f6", created: "2026-09-01", used: "2026-10-30", active: false },
];
const keyColumns: CollectionColumn<Key>[] = [
	{
		id: "name",
		header: "Kunci",
		cell: (key) => (
			<span className="flex flex-col gap-0.5">
				<span className="font-medium">{key.name}</span>
				<span className="font-mono text-xs text-muted-foreground">{key.prefix}…</span>
			</span>
		),
	},
	{ id: "created", header: "Dibuat", inlineLabel: true, cell: (key) => day(key.created) },
	{ id: "used", header: "Terakhir dipakai", inlineLabel: true, cell: (key) => (key.used ? day(key.used) : "Belum pernah") },
	{
		id: "status",
		header: "Status",
		cell: (key) => <Badge variant={key.active ? "success-light" : "outline"}>{key.active ? "Aktif" : "Dicabut"}</Badge>,
	},
	{
		id: "actions",
		header: "Tindakan",
		slot: "actions",
		cell: (key) =>
			key.active ? (
				<Button type="button" variant="outline" size="sm">
					Cabut
				</Button>
			) : null,
	},
];

const narrow = [(Story: () => React.ReactNode) => <div className="max-w-md">{Story()}</div>];
const meta = {
	title: "UI/Collection",
	component: Collection,
	parameters: { layout: "padded" },
	render: () => <SortablePayments />,
	args: { label: "Pembayaran", items: [], columns: [], getKey: () => "" },
	argTypes: { items: { control: false }, columns: { control: false }, getKey: { control: false }, getLink: { control: false } },
} satisfies Meta<typeof Collection>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Transactions compare down a column: a table, the whole row opens the payment, Jumlah and Dibuat sort. */
export const Table: Story = {};
/** Under 42rem of its own width the same definition stacks: title, amount top right, meta below. */
export const TableNarrow: Story = { decorators: narrow };
/** Records without a page do not look clickable; their controls stay usable. Ambiguous dates carry a label when stacked. */
export const TableWithActions: Story = {
	render: () => <Collection label="API key" items={keys} columns={keyColumns} getKey={(key) => key.id} />,
};
export const TableWithActionsNarrow: Story = { ...TableWithActions, decorators: narrow };
/** Records with a face: the same list as cards, several per row, like simple ID cards; the whole card opens the customer. */
export const Cards: Story = {
	render: () => (
		<Collection
			label="Pelanggan"
			layout="cards"
			items={customers}
			columns={customerColumns}
			getKey={(customer) => customer.id}
			getLink={(customer) => <a href={`#${customer.id}`} />}
		/>
	),
};
export const CardsNarrow: Story = { ...Cards, decorators: narrow };
