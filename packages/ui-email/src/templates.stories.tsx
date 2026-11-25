import type { Meta, StoryObj } from "@storybook/react";
import verify from "../previews/verify_email.mjs";
import setup from "../previews/setup_password.mjs";
import reset from "../previews/reset_password.mjs";

const previews = { verify_email: verify, setup_password: setup, reset_password: reset };
function EmailPreview({ purpose }: { purpose: keyof typeof previews }) {
	return (
		<iframe
			title="Pratinjau email transaksional"
			sandbox=""
			tabIndex={-1}
			srcDoc={previews[purpose]}
			style={{ width: "100%", height: 850, border: 0, pointerEvents: "none" }}
		/>
	);
}
const meta = {
	title: "Email/Templates/Transactional",
	component: EmailPreview,
	parameters: { layout: "fullscreen" },
	args: { purpose: "verify_email" },
} satisfies Meta<typeof EmailPreview>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Verification: Story = {};
export const FirstPassword: Story = { args: { purpose: "setup_password" } };
export const ResetPassword: Story = { args: { purpose: "reset_password" } };
