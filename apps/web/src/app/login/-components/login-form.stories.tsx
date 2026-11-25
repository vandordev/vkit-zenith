import type { Meta, StoryObj } from "@storybook/react";
import { LoginForm } from "./login-form";
import { noop, useSubmitCycle } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/Login",
	component: LoginForm,
	parameters: { layout: "fullscreen" },
	args: { status: { kind: "idle" }, onSubmit: noop },
	argTypes: { onSubmit: { control: false }, status: { control: false } },
} satisfies Meta<typeof LoginForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const LongEmail: Story = { args: { defaultEmail: "pelanggan.dengan.alamat.email.panjang@example.test" } };
export const Pending: Story = { args: { status: { kind: "pending" }, defaultEmail: "rani@example.test" } };
export const InvalidInput: Story = {
	args: { status: { kind: "error", message: "Periksa isian Anda.", fields: { email: "Format email tidak valid." } } },
};
export const InvalidCredentials: Story = {
	args: {
		status: { kind: "error", message: "Email atau kata sandi salah. Periksa isian lalu coba lagi." },
		defaultEmail: "rani@example.test",
	},
};
export const RateLimited: Story = { args: { status: { kind: "rate-limited", retryAfterSeconds: 60 } } };
export const Unavailable: Story = { args: { status: { kind: "unavailable" } } };
/** Submit: password clears when the result arrives, email stays. */
export const ErrorKeepsEmail: Story = {
	render: function Preview(args) {
		const [status, submit] = useSubmitCycle({ kind: "error", message: "Email atau kata sandi salah." });
		return <LoginForm {...args} status={status} onSubmit={submit} />;
	},
};
