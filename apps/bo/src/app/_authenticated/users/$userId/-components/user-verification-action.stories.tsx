import type { Meta, StoryObj } from "@storybook/react";
import { UserVerificationAction } from "./user-verification-action";
import { CustomerActionPreview } from "../../-preview";
import { customer, verifiedCustomer } from "../../-fixtures";
import { noop } from "../../../organizations/-fixtures";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
const meta = {
	title: "BO/Pages/Users/VerificationAction",
	component: UserVerificationAction,
	parameters: { layout: "fullscreen" },
	args: {
		record: customer,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1UserGetBody | undefined> => ({
			...customer,
			revision: "2".repeat(64),
			email: "replacement@example.test",
		}),
	},
	render: (args) => (
		<CustomerActionPreview title="Verifikasi email manual" pending={args.pending}>
			{(onDirty, onSubmit) => <UserVerificationAction {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof UserVerificationAction>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const AlreadyVerified: Story = { args: { record: verifiedCustomer } };
export const DisabledUnverified: Story = {
	args: { record: { ...customer, status: "disabled", canChangeStatus: true, statusBlockedReason: "" } },
};
export const Admin: Story = {
	args: { record: { ...customer, canVerifyEmail: false, verificationBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } },
};
export const Pending: Story = { args: { pending: true } };
export const StaleEmail: Story = {
	args: { feedback: { kind: "error", message: "Email berubah. Muat ulang lalu periksa alamat saat ini." } },
};
export const VerifiedDuringConflict: Story = {
	args: {
		...meta.args,
		feedback: { kind: "error", message: "Revisi berubah." },
		onReload: async () => ({ ...verifiedCustomer, id: customer.id }),
	},
};
