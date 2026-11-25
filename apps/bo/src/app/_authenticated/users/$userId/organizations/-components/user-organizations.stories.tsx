import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../../-fixtures";
import { paged } from "../../../../organizations/$organizationId/-setup-fixtures";
import { userMemberships } from "../../../-fixtures";
import { UserOrganizations } from "./user-organizations";

const meta = {
	title: "BO/Pages/Users/Organizations",
	component: UserOrganizations,
	parameters: { layout: "fullscreen" },
	args: {
		memberships: userMemberships,
		state: "ready",
		organizationLink: (row) => <a href={`#organization-${row.organizationId}-${row.id}`} />,
		onPage: noop,
		onRetry: noop,
	},
	argTypes: { organizationLink: { control: false }, memberships: { control: false } },
} satisfies Meta<typeof UserOrganizations>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Pending: Story = { args: { memberships: undefined, state: "pending" } };
export const Error: Story = { args: { memberships: undefined, state: "error", error: "Keanggotaan tidak dapat dimuat." } };
export const NoMemberships: Story = {
	args: { memberships: { items: [], pagination: paged(0), summary: { total: "0", active: "0", removed: "0", owners: "0" } } },
};
export const LaterPageOutOfRange: Story = {
	args: { memberships: { ...userMemberships, items: [], pagination: paged(0, 9, 25, "60") } },
};
export const LongName: Story = {
	args: {
		memberships: {
			...userMemberships,
			items: userMemberships.items.map((row) => ({
				...row,
				organizationName: "Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara",
				organizationSlug: "koperasi-pengembangan-pendidikan-operasional-lintas-wilayah-nusantara",
			})),
		},
	},
};
export const Paging: Story = {
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return <UserOrganizations {...args} memberships={{ ...userMemberships, pagination: paged(2, page, 25, "60") }} onPage={setPage} />;
	},
};
