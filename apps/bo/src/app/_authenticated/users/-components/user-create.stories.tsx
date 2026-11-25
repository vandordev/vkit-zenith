import type { Meta, StoryObj } from "@storybook/react";
import { UserCreate } from "./user-create";
import { CustomerActionPreview } from "../-preview";
import { noop } from "../../organizations/-fixtures";
const meta = {
	title: "BO/Pages/Users/Create",
	component: UserCreate,
	parameters: { layout: "fullscreen" },
	args: { onDirty: noop, onSubmit: noop },
	render: (args) => (
		<CustomerActionPreview title="Tambah pelanggan" pending={args.pending}>
			{(onDirty, onSubmit) => <UserCreate {...args} onDirty={onDirty} onSubmit={onSubmit} />}
		</CustomerActionPreview>
	),
} satisfies Meta<typeof UserCreate>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Email sudah digunakan. Akun tidak digabung dan tidak dipilih otomatis." } },
};
