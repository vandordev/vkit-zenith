import type { Meta, StoryObj } from "@storybook/react";
import { VerifyEmailConfirm } from "./verify-email-confirm";
import { noop } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/VerifyEmail",
	component: VerifyEmailConfirm,
	parameters: { layout: "fullscreen" },
	args: { state: { kind: "ready", expiresAt: "2026-11-10T08:30:00Z" }, confirm: { kind: "idle" }, onConfirm: noop },
	argTypes: { onConfirm: { control: false }, state: { control: false }, confirm: { control: false } },
} satisfies Meta<typeof VerifyEmailConfirm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Ready: Story = {};
export const Loading: Story = { args: { state: { kind: "loading" } } };
export const Confirming: Story = { args: { confirm: { kind: "pending" } } };
export const ConfirmFailed: Story = { args: { confirm: { kind: "error", message: "Verifikasi gagal. Coba lagi." } } };
export const Unavailable: Story = { args: { confirm: { kind: "unavailable" } } };
export const CrossBrowser: Story = { args: { state: { kind: "cross-browser" } } };
export const LinkUnavailable: Story = { args: { state: { kind: "outcome", outcome: "unavailable" } } };
export const CompletedSignedIn: Story = { args: { state: { kind: "outcome", outcome: "completed-signed-in" } } };
export const CompletedLogin: Story = { args: { state: { kind: "outcome", outcome: "completed-login" } } };
