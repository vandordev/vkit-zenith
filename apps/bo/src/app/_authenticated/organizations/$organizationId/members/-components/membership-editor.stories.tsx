import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer, Page } from "../../../../../-components/common";
import { MembershipEditor } from "./membership-editor";
import { membership, ordinaryMember, removedMember, soleOwner } from "../../-fixtures";
import { noop } from "../../../-fixtures";
const meta = {
	title: "BO/Pages/Organizations/MemberEditor",
	component: MembershipEditor,
	parameters: { layout: "fullscreen" },
	args: {
		record: membership,
		onRetry: noop,
		onRemove: noop,
		onRestore: noop,
		onOwner: noop,
		userLink: <a href="#user">Lihat pelanggan</a>,
	},
	render: function Preview(args) {
		const [open, setOpen] = useState(false);
		const [message, setMessage] = useState("");
		return (
			<Page title="Koperasi Nusantara">
				<Button onClick={() => setOpen(true)}>Kelola anggota</Button>
				{message && <p role="status">{message}</p>}
				<EditorDrawer title="Keanggotaan pelanggan" open={open} pending={args.pending} onClose={() => setOpen(false)}>
					{() => (
						<MembershipEditor
							{...args}
							onOwner={() => {
								setOpen(false);
								setMessage("Editor ditutup sebelum konfirmasi kepemilikan dibuka.");
							}}
							onRemove={() => setMessage("Penghapusan anggota biasa diminta")}
							onRestore={() => setMessage("Pemulihan keanggotaan yang sama diminta")}
						/>
					)}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof MembershipEditor>;
export default meta;
type Story = StoryObj<typeof meta>;
export const TwoOwners: Story = {};
export const SoleOwner: Story = { args: { record: soleOwner } };
export const Ordinary: Story = { args: { record: ordinaryMember } };
export const Removed: Story = { args: { record: removedMember } };
export const DisabledOwnerWithReplacement: Story = { args: { record: { ...membership, userStatus: "disabled" } } };
export const DisabledOrdinary: Story = {
	args: {
		record: {
			...ordinaryMember,
			userStatus: "disabled",
			canChangeRole: false,
			roleBlockedReason: "Menambahkan kepemilikan membutuhkan pengguna aktif.",
		},
	},
};
export const DisabledRemoved: Story = {
	args: { record: { ...removedMember, userStatus: "disabled", canRestore: false, restoreBlockedReason: "Pengguna dinonaktifkan." } },
};
export const Admin: Story = {
	args: { record: { ...membership, canChangeRole: false, roleBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } },
};
export const Suspended: Story = { args: { record: { ...ordinaryMember, organizationStatus: "suspended" } } };
export const CrossScopeUnavailable: Story = { args: { record: undefined, error: "Keanggotaan tidak tersedia dalam organisasi ini." } };
export const StaleDetail: Story = { args: { loading: true, error: "Detail belum dapat diperbarui. Muat ulang sebelum bertindak." } };
export const Pending: Story = { args: { record: ordinaryMember, pending: true } };
