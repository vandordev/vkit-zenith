import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { EditForm, Page } from "./common";

const meta = {
	title: "BO/Components/EditForm",
	component: EditForm,
	parameters: { layout: "fullscreen" },
	args: {
		fields: [
			{ name: "email", label: "Email staf", type: "email" },
			{ name: "password", label: "Kata sandi", type: "password", help: "Gunakan kata sandi pratinjau, bukan kredensial asli." },
			{
				name: "role",
				label: "Peran",
				options: [
					{ value: "admin", label: "Admin" },
					{ value: "superadmin", label: "Superadmin" },
				],
			},
			{ name: "key", label: "Kode tetap", readOnly: true },
		],
		initial: { email: "preview@example.invalid", role: "admin", key: "preview" },
		onSubmit: () => {},
	},
	argTypes: { onSubmit: { control: false }, onDirty: { control: false }, fields: { control: false } },
	render: function FormPreview(args) {
		const [dirty, setDirty] = useState(false);
		const [submitted, setSubmitted] = useState(false);
		return (
			<Page title="Pengaturan staf" description="Pratinjau formulir lokal.">
				<div className="max-w-xl">
					<EditForm {...args} onDirty={setDirty} onSubmit={() => setSubmitted(true)} />
					<p role="status" className="mt-4 text-sm">
						{submitted ? "Formulir dikirim; kata sandi dikosongkan." : dirty ? "Perubahan belum disimpan." : "Belum ada perubahan."}
					</p>
				</div>
			</Page>
		);
	},
} satisfies Meta<typeof EditForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const Error: Story = { args: { feedback: { kind: "error", message: "Perubahan belum disimpan. Periksa isian lalu coba lagi." } } };
export const Success: Story = { args: { feedback: { kind: "success", message: "Perubahan disimpan." } } };
