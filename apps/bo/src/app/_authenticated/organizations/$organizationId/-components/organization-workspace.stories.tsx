import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { OrganizationWorkspace } from "./organization-workspace";
import { organization, noop } from "../../-fixtures";
const workspaceOrganization = { ...organization, memberCount: "4", activeMemberCount: "3", activeOwnerCount: "2" };
const meta = {
	title: "BO/Pages/Organizations/Workspace",
	component: OrganizationWorkspace,
	parameters: { layout: "fullscreen" },
	args: {
		record: workspaceOrganization,
		backLink: <a href="#organizations">Kembali ke organisasi</a>,
		onRetry: noop,
		onEdit: noop,
		onStatus: noop,
		onOwners: noop,
	},
	render: function Preview(args) {
		const [message, setMessage] = useState("");
		return (
			<OrganizationWorkspace
				{...args}
				onEdit={() => setMessage("Editor profil diminta")}
				onStatus={() => setMessage("Konfirmasi status diminta")}
				onOwners={() => setMessage("Daftar pemilik diminta")}
			>
				{message && <p role="status">{message}</p>}
			</OrganizationWorkspace>
		);
	},
} satisfies Meta<typeof OrganizationWorkspace>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Admin: Story = {
	args: {
		record: { ...workspaceOrganization, canChangeStatus: false, statusBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." },
	},
};
export const Suspended: Story = { args: { record: { ...workspaceOrganization, status: "suspended" } } };
export const Pending: Story = { args: { record: undefined, loading: true } };
export const Unavailable: Story = { args: { record: undefined, error: "Organisasi tidak tersedia." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan organisasi gagal." } };
export const LongContent: Story = {
	args: {
		record: { ...workspaceOrganization, name: "Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara" },
	},
};
export const EmptyOrganization: Story = {
	args: { record: { ...workspaceOrganization, memberCount: "0", activeMemberCount: "0", activeOwnerCount: "0" } },
};
