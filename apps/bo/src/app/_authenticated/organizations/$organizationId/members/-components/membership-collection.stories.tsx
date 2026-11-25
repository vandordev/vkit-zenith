import type { Meta, StoryObj } from "@storybook/react";
import { Page } from "../../../../../-components/common";
import { MembershipCollection } from "./membership-collection";
import { memberships, membership } from "../../-fixtures";
import { noop } from "../../../-fixtures";
const meta = {
	title: "BO/Pages/Organizations/Memberships",
	component: MembershipCollection,
	parameters: { layout: "fullscreen" },
	args: {
		data: memberships,
		filters: { q: "", status: "", ownership: "" },
		onFilter: noop,
		onPage: noop,
		onRetry: noop,
		onAdd: noop,
		onSelect: noop,
		userLink: (row) => <a href={`#user-${row.userId}`}>{row.userName}</a>,
	},
	decorators: [
		(Story) => (
			<Page title="Koperasi Nusantara">
				<Story />
			</Page>
		),
	],
} satisfies Meta<typeof MembershipCollection>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Pending: Story = { args: { data: undefined, loading: true } };
export const Error: Story = { args: { data: undefined, error: "Anggota tidak dapat dimuat." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan anggota gagal." } };
export const Empty: Story = {
	args: {
		data: { items: [], pagination: { page: 1, pageSize: 25, total: "0" }, summary: { total: "0", active: "0", removed: "0", owners: "0" } },
	},
};
export const NoMatches: Story = {
	args: {
		filters: { q: "Tidak ada", status: "removed", ownership: "member" },
		data: { ...memberships, items: [], pagination: { page: 1, pageSize: 25, total: "0" } },
	},
};
export const OwnersContinuation: Story = {
	args: {
		filters: { q: "", status: "", ownership: "owner" },
		data: {
			...memberships,
			items: [membership],
			pagination: { page: 1, pageSize: 1, total: "2" },
			summary: { total: "15", active: "12", removed: "3", owners: "2" },
		},
	},
};
export const LongContent: Story = {
	args: {
		data: {
			...memberships,
			items: [
				{
					...membership,
					userName: "Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah",
					userEmail: "penanggung-jawab-operasional-pelanggan-lintas-wilayah@example.test",
				},
			],
		},
	},
};
