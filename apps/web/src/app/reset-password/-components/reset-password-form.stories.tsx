import type { Meta, StoryObj } from "@storybook/react";
import { ResetPasswordForm } from "./reset-password-form";
import { noop, useSubmitCycle } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/ResetPassword",
	component: ResetPasswordForm,
	parameters: { layout: "fullscreen" },
	args: { state: { kind: "ready", expiresAt: "2026-11-10T08:30:00Z" }, form: { kind: "idle" }, onSubmit: noop },
	argTypes: { onSubmit: { control: false }, state: { control: false }, form: { control: false } },
} satisfies Meta<typeof ResetPasswordForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Loading: Story = { args: { state: { kind: "loading" } } };
export const Pending: Story = { args: { form: { kind: "pending" } } };
export const InvalidInput: Story = {
	args: { form: { kind: "error", message: "Periksa isian Anda.", fields: { password: "Kata sandi terlalu pendek." } } },
};
export const Unavailable: Story = { args: { form: { kind: "unavailable" } } };
export const CrossBrowser: Story = { args: { state: { kind: "cross-browser" } } };
export const LinkUnavailable: Story = { args: { state: { kind: "outcome", outcome: "unavailable" } } };
export const CompletedLogin: Story = { args: { state: { kind: "outcome", outcome: "completed-login" } } };
/** Submit: password clears when the result arrives. */
export const UnavailableClearsPassword: Story = {
	render: function Preview(args) {
		const [form, submit] = useSubmitCycle({ kind: "unavailable" });
		return <ResetPasswordForm {...args} form={form} onSubmit={submit} />;
	},
};
