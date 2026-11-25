import type { Meta, StoryObj } from "@storybook/react";
import { StaffEditor } from "./staff-editor";
import { Page } from "../../../-components/common";
import { noop, staff } from "../../../-fixtures";

const meta = {
	title: "BO/Pages/Staff/Editor",
	component: StaffEditor,
	parameters: { layout: "fullscreen" },
	args: { initial: { displayName: staff.displayName, email: staff.email, role: "admin", status: "active" }, onDirty: noop, onSubmit: noop },
	argTypes: { initial: { control: false }, onDirty: { control: false }, onSubmit: { control: false } },
	decorators: [
		(Story) => (
			<Page title="Kelola staf">
				<Story />
			</Page>
		),
	],
} satisfies Meta<typeof StaffEditor>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Create: Story = { args: { creating: true } };
export const PasswordReplacement: Story = { args: { replacing: true } };
export const LastSuperadminConflict: Story = {
	args: {
		replacing: true,
		feedback: {
			kind: "error",
			message: "Perubahan ditolak. Setidaknya satu Superadmin aktif harus tetap tersedia. Muat ulang data sebelum mencoba lagi.",
		},
	},
};
