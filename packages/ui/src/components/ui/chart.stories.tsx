import type { Meta, StoryObj } from "@storybook/react";
import { AreaTrend, BarList, Donut, StackedColumns } from "./chart";

const days = Array.from({ length: 30 }, (_, i) => ({ label: `${i + 1} Nov`, value: (2 + ((i * 7) % 5)) * 1_400_000 }));
const rp = (value: number) => `Rp${new Intl.NumberFormat("id-ID").format(value)}`;

const meta = { title: "UI/Chart", parameters: { layout: "padded" } } satisfies Meta;
export default meta;
type Story = StoryObj<typeof meta>;

export const Area: Story = {
	render: () => (
		<AreaTrend
			data={days}
			label="Pendapatan harian"
			formatValue={rp}
			formatTick={(value) => new Intl.NumberFormat("id-ID", { notation: "compact" }).format(value)}
		/>
	),
};

export const Bars: Story = {
	render: () => (
		<BarList
			label="Pembayaran per status"
			formatValue={String}
			data={[
				{ key: "succeeded", label: "Berhasil", value: 1284, color: "var(--success)" },
				{ key: "awaiting", label: "Perlu konfirmasi", value: 3, color: "var(--warning)" },
				{ key: "failed", label: "Gagal", value: 58, color: "var(--destructive)" },
			]}
		/>
	),
};

export const DonutShare: Story = {
	render: () => (
		<Donut
			label="Porsi pembayaran per metode"
			formatValue={String}
			center={{ value: "1.328", caption: "pembayaran" }}
			data={[
				{ key: "qris", label: "QRIS", value: 612 },
				{ key: "va", label: "Virtual account", value: 388 },
				{ key: "card", label: "Kartu", value: 204 },
				{ key: "other", label: "Lainnya", value: 124 },
			]}
		/>
	),
};

export const Stacked: Story = {
	render: () => (
		<StackedColumns
			label="Transaksi harian"
			formatValue={String}
			series={[
				{ key: "succeeded", label: "Berhasil", color: "var(--success)" },
				{ key: "failed", label: "Gagal", color: "var(--destructive)" },
			]}
			data={Array.from({ length: 30 }, (_, i) => ({ label: `${i + 1} Nov`, succeeded: 30 + ((i * 7) % 5) * 9, failed: 1 + ((i * 3) % 4) }))}
		/>
	),
};
