import type { Meta, StoryObj } from "@storybook/react";
import { CheckEmailPanel } from "./check-email-panel";
import { noop } from "../../-story-helpers";

const meta = {
	title: "Web/Pages/CheckEmail",
	component: CheckEmailPanel,
	parameters: { layout: "fullscreen" },
	args: {
		state: { kind: "restricted", email: "rani@example.test", displayName: "Rani Wulandari" },
		resend: { kind: "idle" },
		logout: { kind: "idle" },
		onResend: noop,
		onLogout: noop,
	},
	argTypes: {
		onResend: { control: false },
		onLogout: { control: false },
		state: { control: false },
		resend: { control: false },
		logout: { control: false },
	},
} satisfies Meta<typeof CheckEmailPanel>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Restricted: Story = {};
export const LongNameAndEmail: Story = {
	args: {
		state: {
			kind: "restricted",
			email: "nama.yang.sangat.panjang.sekali.untuk.menguji.tata.letak@subdomain.contoh-perusahaan-panjang.example.test",
			displayName: "Raden Mas Adipati Ario Suryo Notokusumo Hadiningrat Kusumodiningrat",
		},
	},
};
export const ResendPending: Story = { args: { resend: { kind: "pending" } } };
export const ResendSent: Story = { args: { resend: { kind: "sent" } } };
export const ResendLimited: Story = { args: { resend: { kind: "rate-limited", retryAfterSeconds: 60 } } };
export const ResendUnavailable: Story = { args: { resend: { kind: "unavailable" } } };
export const LogoutPending: Story = { args: { logout: { kind: "pending" } } };
export const LogoutFailed: Story = {
	args: { logout: { kind: "error", message: "Gagal keluar. Anda masih masuk; coba lagi." } },
};
export const Generic: Story = { args: { state: { kind: "generic" } } };
