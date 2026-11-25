import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { Page } from "../../../-components/common";
import { StaffGovernance } from "./staff-governance";
import { governanceStaff } from "../-fixtures";
const meta = {
	title: "BO/Pages/Staff/Governance",
	component: StaffGovernance,
	parameters: { layout: "fullscreen" },
	args: { staff: governanceStaff, onDirty: () => {}, onSubmit: () => {} },
	render: function Preview(args) {
		const [captured, setCaptured] = useState("");
		const [background, setBackground] = useState(false);
		const [retried, setRetried] = useState(false);
		return (
			<Page title="Kelola staf">
				<StaffGovernance
					{...args}
					staff={background && args.staff ? { ...args.staff, displayName: "Nama latar", revision: "b".repeat(64) } : args.staff}
					loadError={retried ? undefined : args.loadError}
					onRetry={() => {
						setRetried(true);
						setCaptured("Muat ulang diminta");
					}}
					onReload={async () => ({ ...governanceStaff, revision: "b".repeat(64) })}
					onSubmit={(_values, revision, replacing) =>
						setCaptured(`Pengajuan lokal · ${replacing ? "kata sandi" : "profil"} · revisi ${revision ?? "baru"}`)
					}
				/>
				<Button variant="ghost" onClick={() => setBackground(true)}>
					Pembaruan latar
				</Button>
				<output className="break-all" aria-label="Hasil lokal">
					{captured}
				</output>
			</Page>
		);
	},
} satisfies Meta<typeof StaffGovernance>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Create: Story = { args: { staff: undefined, creating: true } };
export const Pending: Story = { args: { pending: true } };
export const LastSuperadminConflict: Story = {
	args: { feedback: { kind: "error", message: "Setidaknya satu Superadmin aktif harus tetap tersedia." } },
};
export const ExpiredConfirmation: Story = {
	args: { feedback: { kind: "error", message: "Konfirmasi kata sandi kedaluwarsa. Masukkan kembali untuk pengajuan baru." } },
};
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Staf berubah di sesi lain. Muat ulang revisi sebelum mencoba kembali." } },
};
export const RetainedError: Story = { args: { loadError: "Pembaruan staf gagal." } };
