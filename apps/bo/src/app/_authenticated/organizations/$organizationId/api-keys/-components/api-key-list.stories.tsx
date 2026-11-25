import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { noop } from "../../../../../-fixtures";
import { apiKeys, paged } from "../../-setup-fixtures";
import { ApiKeyList } from "./api-key-list";

const meta = {
	title: "BO/Pages/Organizations/ApiKeys",
	component: ApiKeyList,
	parameters: { layout: "fullscreen" },
	args: { keys: apiKeys, state: "ready", pagination: paged(apiKeys.length), onPage: noop, onRetry: noop, onRevoke: noop },
	argTypes: { keys: { control: false } },
} satisfies Meta<typeof ApiKeyList>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Empty: Story = { args: { keys: [], pagination: paged(0) } };
export const Pending: Story = { args: { keys: [], state: "pending", pagination: undefined } };
export const Error: Story = { args: { keys: [], state: "error", error: "Data dummy gagal dimuat", pagination: undefined } };
export const RevokedKeysOnly: Story = { args: { keys: apiKeys.filter((key) => key.status !== "active") } };
export const LaterPage: Story = { args: { pagination: paged(apiKeys.length, 2, 25, "60"), keys: apiKeys.slice(0, 2) } };
export const LaterPageOutOfRange: Story = { args: { pagination: paged(0, 9, 25, "60"), keys: [] } };
export const Paging: Story = {
	args: { pagination: paged(25, 1, 25, "60") },
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return <ApiKeyList {...args} pagination={paged(25, page, 25, "60")} onPage={setPage} />;
	},
};
