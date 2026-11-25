import type { Meta, StoryObj } from "@storybook/react";
import { memberships } from "../../-fixtures";
import { noop } from "../../../-fixtures";
import { MembersPage } from "./members-page";
import { MembershipCollection } from "./membership-collection";

const collection = (
	<MembershipCollection
		data={memberships}
		filters={{ q: "", status: "", ownership: "" }}
		onFilter={noop}
		onPage={noop}
		onRetry={noop}
		onAdd={noop}
		onSelect={noop}
		userLink={(row) => <a href={`#user-${row.userId}`}>{row.userName}</a>}
	/>
);
const meta = {
	title: "BO/Pages/Organizations/Members",
	component: MembersPage,
	parameters: { layout: "fullscreen" },
	args: { onRetry: noop, backLink: <a href="#organizations">Kembali ke daftar asal</a>, children: collection },
	argTypes: { children: { control: false }, backLink: { control: false } },
} satisfies Meta<typeof MembersPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const OrganizationUnavailable: Story = { args: { error: "Organisasi tidak tersedia." } };
