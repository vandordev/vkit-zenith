import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer, Page } from "../../../../../-components/common";
import { OwnerConfirmation } from "./owner-confirmation";
import { membership, ordinaryMember, soleOwner, removedMember } from "../../-fixtures";
import { noop } from "../../../-fixtures";
import type { V1OrganizationMembershipGetBody } from "@bo/lib/api/generated/types.gen";
const meta = {
	title: "BO/Pages/Organizations/OwnerConfirmation",
	component: OwnerConfirmation,
	parameters: { layout: "fullscreen" },
	args: {
		record: membership,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1OrganizationMembershipGetBody | undefined> => ({ ...membership, revision: "2".repeat(64) }),
	},
	render: function Preview(args) {
		const [open, setOpen] = useState(false);
		const [message, setMessage] = useState("");
		return (
			<Page title="Koperasi Nusantara">
				<Button onClick={() => setOpen(true)}>Ubah peran</Button>
				{message && <p role="status">{message}</p>}
				<EditorDrawer title="Ubah peran" open={open} pending={args.pending} onClose={() => setOpen(false)}>
					{(onDirty) => (
						<OwnerConfirmation
							{...args}
							onDirty={onDirty}
							onSubmit={(body) =>
								setMessage(`Peran ${body.role} dikirim. Kata sandi tidak ditampilkan atau dipertahankan; pratinjau lokal.`)
							}
						/>
					)}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof OwnerConfirmation>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Demotion: Story = {};
export const Promotion: Story = { args: { record: ordinaryMember } };
export const SoleOwner: Story = { args: { record: soleOwner } };
export const DisabledOwnerWithReplacement: Story = { args: { record: { ...membership, userStatus: "disabled" } } };
export const Removed: Story = { args: { record: removedMember } };
export const Suspended: Story = { args: { record: { ...membership, organizationStatus: "suspended" } } };
export const Admin: Story = {
	args: { record: { ...membership, canChangeRole: false, roleBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } },
};
export const Pending: Story = { args: { pending: true } };
export const LastOwnerConflict: Story = {
	args: { feedback: { kind: "error", message: "Organisasi harus tetap memiliki pemilik aktif lain." }, onReload: async () => soleOwner },
};
export const StaleConflict: Story = {
	args: { feedback: { kind: "error", message: "Keanggotaan berubah. Muat ulang dan konfirmasi kembali." } },
};
export const UnavailableReload: Story = {
	args: { ...meta.args, feedback: { kind: "error", message: "Keanggotaan berubah." }, onReload: async () => undefined },
};
export const LongContent: Story = {
	args: {
		record: {
			...membership,
			userName: "Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah",
			organizationName: "Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara",
		},
	},
};
