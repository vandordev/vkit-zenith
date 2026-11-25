import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { LoginForm } from "./login-form";
import { noop } from "../../-fixtures";

const meta = {
	title: "BO/Pages/Auth/Login",
	component: LoginForm,
	parameters: { layout: "fullscreen" },
	args: { onSubmit: noop },
	argTypes: { onSubmit: { control: false } },
	render: function LoginPreview(args) {
		const [submitted, setSubmitted] = useState(false);
		return (
			<>
				<LoginForm
					{...args}
					onSubmit={(values) => {
						args.onSubmit(values);
						setSubmitted(true);
					}}
				/>
				{submitted && (
					<p role="status" className="fixed inset-x-4 bottom-4 text-center text-sm">
						Formulir dikirim. Pratinjau lokal, tanpa login ke API.
					</p>
				)}
			</>
		);
	},
} satisfies Meta<typeof LoginForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const InvalidCredentials: Story = {
	args: { feedback: { kind: "error", message: "Email atau kata sandi salah. Periksa isian lalu coba lagi." } },
};
export const RateLimited: Story = {
	args: { feedback: { kind: "error", message: "Terlalu banyak percobaan. Tunggu 600 detik sebelum mencoba lagi." } },
};
export const ServiceUnavailable: Story = { args: { feedback: { kind: "error", message: "Layanan tidak tersedia. Coba lagi nanti." } } };
