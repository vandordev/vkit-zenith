import type { Meta, StoryObj } from "@storybook/react";
import { ActivityPage } from "./activity-page";
import { list, actions, noop } from "../-fixtures";
import { ActivityFilters } from "./activity-filters";
const meta = {
	title: "BO/Pages/Activity/Page",
	component: ActivityPage,
	args: {
		data: list,
		actions,
		onPage: noop,
		onReload: noop,
		onReset: noop,
		renderLink: (row) => <a href={`#${row.id}`} />,
		filters: <ActivityFilters value={{}} actions={actions} onChange={noop} onReset={noop} />,
	},
	parameters: { layout: "fullscreen" },
	argTypes: {
		filters: { control: false },
		children: { control: false },
		renderLink: { control: false },
		onPage: { control: false },
		onReload: { control: false },
		onReset: { control: false },
	},
} satisfies Meta<typeof ActivityPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Superadmin: Story = {};
export const Denied: Story = { args: { denied: true, data: undefined, filters: undefined } };
