import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer } from "../../../-components/common";
import { StaffCollection } from "./staff-collection";
import { StaffPage } from "./staff-page";
import { StaffGovernance } from "./staff-governance";
import { staffList, governanceStaff } from "../-fixtures";
const meta = {
	title: "BO/Pages/Staff/Collection",
	component: StaffCollection,
	parameters: { layout: "fullscreen" },
	args: { data: staffList, q: "", status: "", onFilter: () => {}, onPage: () => {}, onReload: () => {}, onOpen: () => {} },
	render: function Preview(args) {
		const [q, setQ] = useState(args.q);
		const [status, setStatus] = useState(args.status);
		const [selected, setSelected] = useState(false);
		const [creating, setCreating] = useState(false);
		const [captured, setCaptured] = useState("");
		return (
			<StaffPage
				{...args}
				q={q}
				status={status}
				action={
					<Button
						onClick={() => {
							setCreating(true);
							setSelected(true);
						}}
					>
						Tambah staf
					</Button>
				}
				onFilter={(q, status) => {
					setQ(q);
					setStatus(status);
				}}
				onPage={(p) => setCaptured(`Halaman ${p}`)}
				onReload={() => setCaptured("Muat ulang diminta")}
				onOpen={() => {
					setCreating(false);
					setSelected(true);
				}}
			>
				<EditorDrawer open={selected} title={creating ? "Tambah staf" : "Ubah staf"} onClose={() => setSelected(false)}>
					{(onDirty) => (
						<StaffGovernance
							key={String(creating)}
							staff={creating ? undefined : governanceStaff}
							creating={creating}
							onDirty={onDirty}
							onSubmit={() => setCaptured("Pengajuan lokal diterima")}
						/>
					)}
				</EditorDrawer>
				{captured && <p role="status">{captured}</p>}
			</StaffPage>
		);
	},
} satisfies Meta<typeof StaffCollection>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Pending: Story = { args: { data: undefined, loading: true } };
export const RetainedPending: Story = { args: { loading: true } };
export const Error: Story = { args: { data: undefined, error: "Staf tidak dapat dimuat." } };
export const RetainedError: Story = { args: { error: "Pembaruan staf gagal." } };
export const Empty: Story = {
	args: {
		data: {
			items: [],
			summary: { total: "0", active: "0", disabled: "0", superadmin: "0", admin: "0" },
			pagination: { page: 1, pageSize: 25, total: 0 },
		},
	},
};
export const NoMatches: Story = {
	args: { q: "Tidak ditemukan", data: { ...staffList, items: [], pagination: { page: 1, pageSize: 25, total: 0 } } },
};
export const Paginated: Story = { args: { data: { ...staffList, pagination: { page: 1, pageSize: 1, total: 42 } } } };
export const LongContent: Story = {
	args: {
		data: {
			...staffList,
			items: staffList.items.map((i) => ({
				...i,
				displayName: "Nadia Prameswari dengan nama sangat panjang untuk pemeriksaan tampilan pada perangkat seluler",
			})),
		},
	},
};
