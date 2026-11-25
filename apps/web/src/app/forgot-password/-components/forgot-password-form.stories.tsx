import type { Meta, StoryObj } from "@storybook/react";
import { ForgotPasswordForm, RequestAccepted } from "./forgot-password-form";
import { noop } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/ForgotPassword",
	component: ForgotPasswordForm,
	parameters: { layout: "fullscreen" },
	args: { status: { kind: "idle" }, onSubmit: noop },
	argTypes: { onSubmit: { control: false }, status: { control: false } },
} satisfies Meta<typeof ForgotPasswordForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { status: { kind: "pending" } } };
export const InvalidInput: Story = {
	args: { status: { kind: "error", message: "Periksa isian Anda.", fields: { email: "Format email tidak valid." } } },
};
export const RateLimited: Story = { args: { status: { kind: "rate-limited", retryAfterSeconds: 300 } } };
export const Unavailable: Story = { args: { status: { kind: "unavailable" } } };
export const Accepted: Story = { render: () => <RequestAccepted /> };
