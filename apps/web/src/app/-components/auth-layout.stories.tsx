import type { Meta, StoryObj } from "@storybook/react";
import { AuthLayout } from "./auth-layout";
import { paths } from "../../lib/paths";

const meta = {
	title: "Web/Layouts/AuthLayout",
	component: AuthLayout,
	parameters: { layout: "fullscreen" },
	args: {
		title: "Masuk ke akun Zenith",
		description: "Gunakan email dan kata sandi Anda.",
		children: <p className="text-sm">Konten formulir.</p>,
		footer: <a href={paths.register}>Belum punya akun? Daftar</a>,
	},
	argTypes: { children: { control: false }, footer: { control: false } },
} satisfies Meta<typeof AuthLayout>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const WithoutFooter: Story = {
	args: { title: "Memeriksa tautan…", description: undefined, footer: undefined },
};
export const LongTitle: Story = {
	args: {
		title: "Verifikasi alamat email Anda untuk melanjutkan pendaftaran akun Zenith di seluruh aplikasi yang terhubung",
		description: "Tautan verifikasi berlaku untuk waktu terbatas dan hanya dapat digunakan satu kali oleh pemilik akun.",
	},
};
