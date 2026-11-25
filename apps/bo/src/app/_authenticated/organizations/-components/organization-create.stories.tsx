import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { OrganizationCreate } from "./organization-create";
import { EditorDrawer, Page } from "../../../-components/common";
import { disabledOwner, owner, ownerChoices, noop } from "../-fixtures";
const meta = {
	title: "BO/Pages/Organizations/Create",
	component: OrganizationCreate,
	parameters: { layout: "fullscreen" },
	args: {
		onDirty: noop,
		onSubmit: noop,
		choices: { data: ownerChoices, value: null, q: "", page: 1, onSelected: noop, onSearch: noop, onPage: noop, onRetry: noop },
	},
	argTypes: { choices: { control: false }, onDirty: { control: false }, onSubmit: { control: false } },
	render: function Preview(args) {
		const [open, setOpen] = useState(false);
		const [q, setQ] = useState("");
		const [page, setPage] = useState(1);
		const [selected, setSelected] = useState(args.choices.value);
		const [message, setMessage] = useState("");
		return (
			<Page
				title="Organisasi"
				action={
					<Button
						onClick={() => {
							setSelected(args.choices.value);
							setQ("");
							setOpen(true);
						}}
					>
						Tambah organisasi
					</Button>
				}
			>
				{message && (
					<p role="status" className="break-words">
						{message}
					</p>
				)}
				<EditorDrawer title="Tambah organisasi" open={open} pending={args.pending} onClose={() => setOpen(false)}>
					{(onDirty) => (
						<OrganizationCreate
							{...args}
							onDirty={onDirty}
							choices={{
								...args.choices,
								value: selected,
								q,
								page,
								onSearch: (next) => {
									setQ(next);
									setPage(1);
								},
								onPage: setPage,
								onSelected: setSelected,
								onRetry: () => setMessage("Muat ulang pilihan diminta"),
								data: args.choices.data
									? {
											...args.choices.data,
											items: args.choices.data.items.filter(
												(item) => !q || `${item.displayName} ${item.email}`.toLowerCase().includes(q.toLowerCase()),
											),
											selected,
										}
									: undefined,
							}}
							onSubmit={(body) => {
								setMessage(
									`Organisasi dikirim: ${body.name}, pemilik ${body.newOwner ? "baru" : "lama"}. Ini pratinjau lokal, bukan penyimpanan API.`,
								);
								if (!args.feedback) setOpen(false);
							}}
						/>
					)}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof OrganizationCreate>;
export default meta;
type Story = StoryObj<typeof meta>;
export const ExistingOwner: Story = {};
export const NewOwner: Story = { args: { initialMode: "new" } };
export const DisabledSelectedOwner: Story = { args: { choices: { ...meta.args.choices, value: disabledOwner } } };
export const DuplicateConflict: Story = {
	args: {
		choices: { ...meta.args.choices, value: owner },
		feedback: { kind: "error", message: "Slug organisasi atau email sudah digunakan. Periksa isian lalu kirim kembali." },
	},
};
export const Pending: Story = { args: { pending: true, choices: { ...meta.args.choices, value: owner } } };
export const PendingNewOwner: Story = { args: { pending: true, initialMode: "new" } };
export const ChoiceFailure: Story = {
	args: { choices: { ...meta.args.choices, data: undefined, error: "Pilihan pelanggan tidak dapat dimuat." } },
};
export const ChoicePending: Story = { args: { choices: { ...meta.args.choices, data: undefined, loading: true } } };
export const LongSelectedOwner: Story = {
	args: {
		choices: {
			...meta.args.choices,
			value: {
				...owner,
				displayName: "Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah",
				email: "penanggung-jawab-operasional-pelanggan-lintas-wilayah@example.test",
			},
		},
	},
};
