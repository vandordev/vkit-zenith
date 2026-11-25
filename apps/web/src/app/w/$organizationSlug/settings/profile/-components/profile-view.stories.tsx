import type { Meta, StoryObj } from "@storybook/react";
import { ProfileView } from "./profile-view";
import { profile } from "../../../-components/-fixtures";

import { delayed, describeError, failing } from "../../../-components/-story-fixtures";
import { organization } from "../../../-components/-fixtures";
import { ShellPreview } from "../../../-components/-shell-preview";

const noop = () => {};
const meta = {
	title: "Web/Pages/Dashboard/Settings/Profile",
	component: ProfileView,
	parameters: { layout: "fullscreen", routerInitialPath: "/w/kopi-senja/settings/profile" },
	args: {
		state: "ready",
		profile,
		access: "allowed",
		describeError,
		onRetry: noop,
		onUploadLogo: (body) => delayed({ contentType: body.contentType, size: body.size }),
		onLogoResult: noop,
	},
	argTypes: { profile: { control: false } },
} satisfies Meta<typeof ProfileView>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Owner, no logo: the organization initial is the fallback in the identity preview, with the dummy upload button. */
export const Default: Story = {};
/** Admin and member share one view: the owner-only note replaces the upload. */
export const Admin: Story = { args: { access: "no-permission" } };
export const Member: Story = { args: { access: "no-permission" } };
/** Suspended organization: the banner explains, nothing can be uploaded. */
export const SuspendedNoWrites: Story = { args: { access: "suspended", profile: { ...profile, status: "suspended" } } };
/** After a dummy upload: only the chosen type and size are noted; no file existed, so the preview still shows the initial. */
export const LogoSimulated: Story = { args: { simulatedLogo: { contentType: "image/png", size: 24576 } } };
export const UploadFails: Story = { args: { onUploadLogo: () => failing(422) } };
export const Mobile: Story = { globals: { viewport: { value: "mobile1", isRotated: false } } };
export const InShell: Story = {
	render: (args) => (
		<ShellPreview org={organization("owner")}>
			<ProfileView {...args} />
		</ShellPreview>
	),
};
export const WithLogo: Story = { args: { profile: { ...profile, hasLogo: true } } };
export const Suspended: Story = { args: { profile: { ...profile, status: "suspended" } } };
export const LongName: Story = {
	args: {
		profile: {
			...profile,
			name: "Perkumpulan Penggiat Kopi Nusantara Tenggara Timur dan Sekitarnya Cabang Utama Sejahtera",
			slug: "perkumpulan-penggiat-kopi-nusantara-tenggara-timur-dan-sekitarnya",
		},
	},
};
export const Pending: Story = { args: { state: "pending", profile: undefined } };
export const Error: Story = {
	args: { state: "error", profile: undefined, errorMessage: "Layanan sedang tidak tersedia. Coba lagi sebentar lagi." },
};
