import type { Meta, StoryObj } from "@storybook/react";
import { UserEmailAction } from "./user-email-action";
import { CustomerActionPreview } from "../../-preview";
import { customer } from "../../-fixtures";
import { noop } from "../../../organizations/-fixtures";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
import { useState } from "react";
import { Button } from "@repo/ui/button";
const meta = {
	title: "BO/Pages/Users/EmailAction",
	component: UserEmailAction,
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
		<CustomerActionPreview title="Ganti email" pending={args.pending}>
			{(onDirty, onSubmit) => <UserEmailAction {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof UserEmailAction>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Admin: Story = {
	args: { record: { ...customer, canChangeEmail: false, emailBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi." } },
};
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Email atau revisi berubah. Muat ulang sebelum konfirmasi ulang." } },
};
export const Duplicate: Story = { args: { feedback: { kind: "error", message: "Email sudah digunakan. Akun tidak digabung." } } };
export const LongEmail: Story = {
	args: { record: { ...customer, email: "penanggung-jawab-operasional-pelanggan-lintas-wilayah@example.test" } },
};
export const ReloadFailure: Story = { args: { ...Conflict.args, onReload: async () => undefined } };
export const IdentityChange: Story = {
	render: function Preview(args) {
		const [record, setRecord] = useState(customer);
		return (
			<CustomerActionPreview title="Ganti email">
				{(onDirty, onSubmit) => (
					<>
						<Button
							variant="outline"
							onClick={() =>
								setRecord({
									...customer,
									id: "00000000-0000-4000-8000-000000000209",
									displayName: "Pelanggan berbeda",
									email: "other@example.test",
								})
							}
						>
							Ganti target pratinjau
						</Button>
						<UserEmailAction {...args} record={record} onDirty={onDirty} onSubmit={onSubmit} />
					</>
				)}
			</CustomerActionPreview>
		);
	},
};
