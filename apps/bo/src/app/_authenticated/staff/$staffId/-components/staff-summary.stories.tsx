import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { governanceStaff, staffList } from "../../-fixtures";
import { StaffSummary } from "./staff-summary";

const noop = () => {};
const meta = {
	title: "BO/Pages/Staff/Summary",
	component: StaffSummary,
	parameters: { layout: "fullscreen" },
	args: {
		staff: governanceStaff,
		backLink: <a href="#staff">Kembali ke daftar staf</a>,
		onRetry: noop,
		onDirty: noop,
		onSubmit: noop,
		onReload: async () => governanceStaff,
	},
	render: function Preview(args) {
		const [captured, setCaptured] = useState("");
		return (
			<>
				<StaffSummary {...args} onSubmit={() => setCaptured("Pengajuan lokal diterima")} />
				{captured && <p role="status">{captured}</p>}
			</>
		);
	},
	argTypes: { staff: { control: false }, backLink: { control: false } },
} satisfies Meta<typeof StaffSummary>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const DisabledAdmin: Story = { args: { staff: staffList.items[1] } };
export const Pending: Story = { args: { staff: undefined } };
export const Error: Story = { args: { staff: undefined, error: "Staf tidak tersedia." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan staf gagal." } };
export const Refreshing: Story = { args: { loading: true } };
export const Saving: Story = { args: { pending: true } };
export const SaveFailure: Story = { args: { feedback: { kind: "error", message: "Revisi staf sudah berubah. Muat ulang revisi staf." } } };
export const Saved: Story = { args: { feedback: { kind: "success", message: "Perubahan berhasil disimpan." } } };
export const LongContent: Story = {
	args: {
		staff: {
			...governanceStaff,
			displayName: "Muhammad Alexander Wijayakusuma Hadiningrat Prawirodirjo Santoso dari Divisi Operasional",
			email: "staf-dengan-alamat-panjang-untuk-pemeriksaan-tata-letak@example.test",
		},
	},
};
