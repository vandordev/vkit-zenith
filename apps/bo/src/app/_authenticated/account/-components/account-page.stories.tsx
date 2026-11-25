import type { Meta, StoryObj } from "@storybook/react";
import { AccountPage } from "./account-page";
import { noop, staff } from "../../../-fixtures";
import { useState } from "react";

const meta = {
	title: "BO/Pages/Account/Profile",
	component: AccountPage,
	parameters: { layout: "fullscreen" },
	args: { staff, onSubmit: noop },
	argTypes: { staff: { control: false }, onSubmit: { control: false } },
	render: function Preview(args) {
		const [submitted, setSubmitted] = useState(false);
		return (
			<>
				<AccountPage {...args} onSubmit={() => setSubmitted(true)} />
				{submitted && <p role="status">Pengajuan lokal diterima. Kata sandi tidak disimpan dalam pratinjau.</p>}
			</>
		);
	},
} satisfies Meta<typeof AccountPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const Invalid: Story = {
	args: {
		feedback: { kind: "error", message: "Periksa batas kata sandi, kesamaan konfirmasi dan pastikan berbeda dari kata sandi saat ini." },
	},
};
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Perubahan tidak dapat diselesaikan. Masukkan kembali konfirmasi sebelum mencoba lagi." } },
};
export const LongIdentity: Story = {
	args: {
		staff: {
			...staff,
			email: "staf-dengan-alamat-panjang-untuk-pemeriksaan-responsif@example.test",
			displayName: "Nadia Prameswari untuk tim operasional seluruh cabang",
		},
	},
};
