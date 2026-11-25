import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer, Page } from "../../../../../-components/common";
import { MembershipAdd } from "./membership-editor";
import { ownerChoices, owner, disabledOwner, noop } from "../../../-fixtures";
const meta = {
	title: "BO/Pages/Organizations/MemberAdd",
	component: MembershipAdd,
	parameters: { layout: "fullscreen" },
	args: {
		onDirty: noop,
		onSubmit: noop,
		choices: { value: null, data: ownerChoices, q: "", page: 1, onSelected: noop, onSearch: noop, onPage: noop, onRetry: noop },
	},
	render: function Preview(args) {
		const [open, setOpen] = useState(false);
		const [value, setValue] = useState(args.choices.value);
		const [q, setQ] = useState("");
		const [message, setMessage] = useState("");
		return (
			<Page title="Koperasi Nusantara">
				<Button onClick={() => setOpen(true)}>Tambah anggota</Button>
				{message && <p role="status">{message}</p>}
				<EditorDrawer title="Tambah anggota" open={open} pending={args.pending} onClose={() => setOpen(false)}>
					{(onDirty) => (
						<MembershipAdd
							{...args}
							onDirty={onDirty}
							choices={{ ...args.choices, value, q, onSelected: setValue, onSearch: setQ }}
							onSubmit={() => {
								setMessage("Anggota biasa dikirim. Pratinjau lokal, bukan penyimpanan API.");
								if (!args.feedback) setOpen(false);
							}}
						/>
					)}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof MembershipAdd>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Selected: Story = { args: { choices: { ...meta.args.choices, value: owner } } };
export const IneligibleSelected: Story = { args: { choices: { ...meta.args.choices, value: disabledOwner } } };
export const ExistingRemoved: Story = {
	args: {
		choices: {
			...meta.args.choices,
			value: { ...owner, eligible: false, disabledReason: "Keanggotaan sudah ada. Pulihkan keanggotaan yang dihapus." },
		},
	},
};
export const Pending: Story = { args: { pending: true, choices: { ...meta.args.choices, value: owner } } };
export const Conflict: Story = {
	args: {
		choices: { ...meta.args.choices, value: owner },
		feedback: { kind: "error", message: "Keanggotaan berubah atau sudah ada. Muat ulang detail." },
	},
};
export const Error: Story = {
	args: { choices: { ...meta.args.choices, data: undefined, error: "Pilihan pelanggan tidak dapat dimuat." } },
};
