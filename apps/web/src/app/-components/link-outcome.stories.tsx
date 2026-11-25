import type { Meta, StoryObj } from "@storybook/react";
import { LinkOutcome } from "./link-outcome";

const meta = {
	title: "Web/Components/LinkOutcome",
	component: LinkOutcome,
	parameters: { layout: "fullscreen" },
	args: { kind: "unavailable" },
} satisfies Meta<typeof LinkOutcome>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Unavailable: Story = {};
export const LoginRequired: Story = { args: { kind: "login-required" } };
export const CompletedLogin: Story = { args: { kind: "completed-login" } };
export const CompletedSignedIn: Story = { args: { kind: "completed-signed-in" } };
