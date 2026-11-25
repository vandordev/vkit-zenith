import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { RegisterForm } from "./register-form";
import { noop, useSubmitCycle } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/Register",
	component: RegisterForm,
	parameters: { layout: "fullscreen" },
	args: { status: { kind: "idle" }, onSubmit: noop },
	argTypes: { onSubmit: { control: false }, status: { control: false }, defaults: { control: false } },
} satisfies Meta<typeof RegisterForm>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = {
	args: {
		status: { kind: "pending" },
		defaults: { displayName: "Rani Wulandari", email: "rani@example.test", telephone: "6281234567890" },
	},
};
export const InvalidInput: Story = {
	args: {
		status: {
			kind: "error",
			message: "Periksa isian Anda.",
			fields: { email: "Format email tidak valid.", password: "Kata sandi terlalu pendek." },
		},
		defaults: { displayName: "Rani Wulandari", email: "rani@example", telephone: "6281234567890" },
	},
};
export const LongName: Story = {
	args: { defaults: { displayName: "Raden Mas Adipati Ario Suryo Notokusumo Hadiningrat Kusumodiningrat dari Surakarta Hadiningrat" } },
};
export const RateLimited: Story = { args: { status: { kind: "rate-limited", retryAfterSeconds: 120 } } };
export const Unavailable: Story = { args: { status: { kind: "unavailable" } } };
/** Submit: password clears when the result arrives, other fields stay. */
export const UnavailableKeepsFields: Story = {
	args: { defaults: { displayName: "Rani Wulandari", email: "rani@example.test", telephone: "6281234567890" } },
	render: function Preview(args) {
		const [status, submit] = useSubmitCycle({ kind: "unavailable" });
		return <RegisterForm {...args} status={status} onSubmit={submit} />;
	},
};
/** Core accepts up to 200 Unicode codepoints; the form must pass the name through intact. */
export const SubmitReportsNameLength: Story = {
	render: function Preview(args) {
		const [received, setReceived] = useState<string>();
		return (
			<>
				<RegisterForm {...args} onSubmit={(values) => setReceived(values.displayName)} />
				{received !== undefined && (
					<p role="status" data-received={received} className="fixed inset-x-4 bottom-4 text-center text-sm">
						Diterima: {[...received].length} karakter. Pratinjau lokal, tanpa API.
					</p>
				)}
			</>
		);
	},
};
