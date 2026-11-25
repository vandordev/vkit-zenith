import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer, Page } from "../../../-components/common";
import { OrganizationProfile } from "./organization-profile";
import { organization, noop } from "../-fixtures";
import type { V1OrganizationGetBody } from "@bo/lib/api/generated/types.gen";
const meta = {
	title: "BO/Pages/Organizations/Profile",
	component: OrganizationProfile,
	parameters: { layout: "fullscreen" },
	args: {
		record: organization,
		onDirty: noop,
		onSubmit: noop,
		onReload: async (): Promise<V1OrganizationGetBody | undefined> => ({ ...organization, revision: "2".repeat(64), name: "Nama terbaru" }),
	},
	argTypes: { onDirty: { control: false }, onSubmit: { control: false }, onReload: { control: false } },
	render: function Preview(args) {
		const [open, setOpen] = useState(false);
		const [message, setMessage] = useState("");
		return (
			<Page title={args.record.name} action={<Button onClick={() => setOpen(true)}>Ubah profil</Button>}>
				{message && <p role="status">{message}</p>}
				<EditorDrawer
					title="Ubah profil organisasi"
					open={open}
					pending={args.pending}
					editorKey={args.record.id}
					onClose={() => setOpen(false)}
				>
					{(onDirty) => (
						<OrganizationProfile
							{...args}
							onDirty={onDirty}
							onSubmit={(body) => {
								setMessage(`Profil ${body.name} dikirim pada revisi ${body.revision.slice(0, 4)}. Pratinjau lokal.`);
								if (!args.feedback) setOpen(false);
							}}
						/>
					)}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof OrganizationProfile>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
export const Conflict: Story = {
	args: { feedback: { kind: "error", message: "Organisasi berubah. Muat ulang revisi sebelum mengirim kembali." } },
};
export const RefreshFailure: Story = { args: { refreshError: "Pembaruan latar belakang gagal." } };
export const ReloadUnavailable: Story = { args: { ...Conflict.args, onReload: async () => undefined } };
export const LongContent: Story = {
	args: {
		record: {
			...organization,
			name: "Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara",
			slug: "koperasi-pengembangan-pendidikan-operasional-lintas-wilayah",
		},
	},
};
