import type { Meta, StoryObj } from "@storybook/react";
import { UserProfile } from "./user-profile";
import { CustomerActionPreview } from "../../-preview";
import { customer } from "../../-fixtures";
import { noop } from "../../../organizations/-fixtures";
import type { V1UserGetBody } from "@bo/lib/api/generated/types.gen";
const meta = {
	title: "BO/Pages/Users/Profile",
	component: UserProfile,
	parameters: { layout: "fullscreen" },
	args: {
		record: customer,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1UserGetBody | undefined> => ({ ...customer, revision: "2".repeat(64), displayName: "Nama terbaru" }),
	},
	render: (args) => (
		<CustomerActionPreview title="Ubah profil" pending={args.pending}>
			{(onDirty, onSubmit) => <UserProfile {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof UserProfile>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = { args: { feedback: { kind: "error", message: "Profil berubah. Muat ulang sebelum mengirim kembali." } } };
export const ReloadFailure: Story = { args: { ...Conflict.args, onReload: async () => undefined } };
