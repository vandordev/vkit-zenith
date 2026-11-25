import type { Meta, StoryObj } from "@storybook/react";

import { QrCode } from "./qr-code";

const meta = {
	argTypes: {
		label: { control: "text" },
		size: { control: { max: 480, min: 96, step: 8, type: "range" } },
		value: { control: "text" },
	},
	args: {
		label: "Kode QR untuk membuka halaman pembayaran",
		size: 192,
		value: "https://checkout.example.test/pay/0f3c1c0e-5a1b-4a52-9c3e-7d2f6a1b8e44",
	},
	component: QrCode,
	parameters: { layout: "centered" },
	title: "UI/QR Code",
} satisfies Meta<typeof QrCode>;

export default meta;
type Story = StoryObj<typeof meta>;
export const Playground: Story = {};
export const Small: Story = { args: { size: 120 } };
export const Large: Story = { args: { size: 320 } };
export const LongValue: Story = { args: { value: `https://checkout.example.test/pay/${"a".repeat(400)}` } };
export const Qris: Story = {
	args: {
		label: "Kode QRIS untuk membayar Rp 25.000",
		value: "00020101021226570011ID.DANA.WWW0118936009153000000015802ID5910Toko Kopi 6013Jakarta Pusat6304ABCD",
	},
};
export const Unicode: Story = { args: { value: "Kopi Susu — Rp 25.000 ✓ 日本語 😀 café" } };
export const Responsive: Story = {
	args: { className: "h-auto w-full max-w-72" },
	render: (args) => <QrCode {...args} />,
};
