import type { Meta, StoryObj } from "@storybook/react";
import { SetPasswordForm } from "./set-password-form";
import { noop, useSubmitCycle } from "../../-story-helpers";

const ready = { kind: "ready", expiresAt: "2026-11-10T08:30:00Z", displayName: "Rani Wulandari" } as const;

const meta = {
	title: "Web/Pages/SetPassword",
	component: SetPasswordForm,
	parameters: { layout: "fullscreen" },
	args: { state: ready, form: { kind: "idle" }, onSubmit: noop },
	argTypes: { onSubmit: { control: false }, state: { control: false }, form: { control: false } },
} satisfies Meta<typeof SetPasswordForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const LongName: Story = {
	args: {
		state: { ...ready, displayName: "Raden Mas Adipati Ario Suryo Notokusumo Hadiningrat Kusumodiningrat dari Surakarta Hadiningrat" },
	},
};
export const Loading: Story = { args: { state: { kind: "loading" } } };
export const Pending: Story = { args: { form: { kind: "pending" } } };
export const InvalidInput: Story = {
	args: {
		form: {
			kind: "error",
			message: "Periksa isian Anda.",
			fields: { telephone: "Nomor telepon tidak valid.", password: "Kata sandi terlalu pendek." },
		},
	},
};
export const Unavailable: Story = { args: { form: { kind: "unavailable" } } };
export const CrossBrowser: Story = { args: { state: { kind: "cross-browser" } } };
export const LinkUnavailable: Story = { args: { state: { kind: "outcome", outcome: "unavailable" } } };
export const CompletedLogin: Story = { args: { state: { kind: "outcome", outcome: "completed-login" } } };
/** Submit: password clears when the result arrives, telephone stays. */
export const UnavailableKeepsTelephone: Story = {
	render: function Preview(args) {
		const [form, submit] = useSubmitCycle({ kind: "unavailable" });
		return <SetPasswordForm {...args} form={form} onSubmit={submit} />;
	},
};
