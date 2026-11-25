import type { Meta, StoryObj } from "@storybook/react";
import { OrganizationStatusAction } from "./organization-status-action";
import { CustomerActionPreview } from "../../users/-preview";
import { organization, noop } from "../-fixtures";
import type { V1OrganizationGetBody } from "@bo/lib/api/generated/types.gen";
const meta = {
	title: "BO/Pages/Organizations/StatusAction",
	component: OrganizationStatusAction,
	parameters: { layout: "fullscreen" },
	args: {
		record: organization,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1OrganizationGetBody | undefined> => ({ ...organization, revision: "2".repeat(64) }),
	},
	render: (args) => (
		<CustomerActionPreview title="Ubah status organisasi" pageTitle="Organisasi" pending={args.pending}>
			{(onDirty, onSubmit) => <OrganizationStatusAction {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof OrganizationStatusAction>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Suspend: Story = {};
export const Reactivate: Story = { args: { record: { ...organization, status: "suspended" } } };
export const Admin: Story = {
	args: { record: { ...organization, canChangeStatus: false, statusBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } },
};
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Organisasi berubah. Muat ulang sebelum konfirmasi ulang." } },
};
