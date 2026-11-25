import type { Meta, StoryObj } from "@storybook/react";
import { MembersList } from "./members-list";
import { members, singleOwner } from "../../../-components/-fixtures";

const noop = () => {};
const meta = {
	title: "Web/Pages/Dashboard/Settings/Members",
	component: MembersList,
	parameters: { layout: "fullscreen" },
	args: { state: "ready", members, onRetry: noop },
	argTypes: { members: { control: false } },
} satisfies Meta<typeof MembersList>;
export default meta;
type Story = StoryObj<typeof meta>;

export const ManyMembers: Story = {};
export const SingleOwner: Story = { args: { members: singleOwner } };
export const LongNames: Story = {
	args: {
		members: [
			{
				name: "Nama Anggota Yang Sangat Panjang Sekali Untuk Menguji Pembungkusan Teks Tabel",
				email: "alamat.email.yang.sangat.panjang.sekali.untuk.uji@contoh-domain-yang-panjang.example.com",
				role: "owner",
				joinedAt: "2026-01-10T03:00:00Z",
			},
		],
	},
};
export const Empty: Story = { args: { members: [] } };
export const Pending: Story = { args: { state: "pending", members: undefined } };
export const Error: Story = {
	args: { state: "error", members: undefined, errorMessage: "Layanan sedang tidak tersedia. Coba lagi sebentar lagi." },
};
