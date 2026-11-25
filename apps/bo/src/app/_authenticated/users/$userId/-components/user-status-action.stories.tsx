import type { Meta, StoryObj } from "@storybook/react";
import { UserStatusAction } from "./user-status-action";
import { CustomerActionPreview } from "../../-preview";
import { customer, verifiedCustomer } from "../../-fixtures";
import { noop } from "../../../organizations/-fixtures";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
const eligible = { ...customer, canChangeStatus: true, statusBlockedReason: "" };
const meta = {
	title: "BO/Pages/Users/StatusAction",
	component: UserStatusAction,
	parameters: { layout: "fullscreen" },
	args: {
		record: eligible,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1UserGetBody | undefined> => ({ ...eligible, revision: "2".repeat(64) }),
	},
	render: (args) => (
		<CustomerActionPreview title="Ubah status pelanggan" pending={args.pending}>
			{(onDirty, onSubmit) => <UserStatusAction {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof UserStatusAction>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Disable: Story = {};
export const Reactivate: Story = { args: { record: verifiedCustomer } };
export const SoleOwner: Story = { args: { record: customer } };
export const Admin: Story = { args: { record: { ...customer, statusBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } } };
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = {
	args: {
		feedback: { kind: "error", message: "Pemilik aktif terakhir tidak dapat dinonaktifkan. Muat ulang." },
		onReload: async () => customer,
	},
};
