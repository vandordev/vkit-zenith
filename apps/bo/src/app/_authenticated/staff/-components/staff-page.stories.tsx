import type { Meta, StoryObj } from "@storybook/react";
import { StaffPage } from "./staff-page";
import { AccessState } from "../../../-components/access-state";
import { staffList } from "../-fixtures";
const meta = {
	title: "BO/Pages/Staff/Page",
	component: StaffPage,
	parameters: { layout: "fullscreen" },
	args: { data: staffList, q: "", status: "", onFilter: () => {}, onPage: () => {}, onReload: () => {}, onOpen: () => {} },
} satisfies Meta<typeof StaffPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Denied: Story = { render: () => <AccessState kind="denied" /> };
