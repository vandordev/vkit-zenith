import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { ActivityCollection } from "./activity-collection";
import { list, actions, noop, event, customerEvent } from "../-fixtures";
const meta = {
	title: "BO/Pages/Activity/Collection",
	component: ActivityCollection,
	args: {
		data: list,
		actions,
		onPage: noop,
		onReload: noop,
		onReset: noop,
		renderLink: (row) => <a href={`#${row.id}`} />,
	},
	parameters: { layout: "padded" },
	argTypes: { renderLink: { control: false }, onPage: { control: false }, onReload: { control: false }, onReset: { control: false } },
} satisfies Meta<typeof ActivityCollection>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {
	render: function Preview(args) {
		const [page, setPage] = useState(1);
		return (
			<>
				<ActivityCollection
					{...args}
					data={args.data ? { ...args.data, pagination: { ...args.data.pagination, page } } : undefined}
					onPage={setPage}
				/>
				<output>Halaman diminta: {page}</output>
			</>
		);
	},
};
export const Empty: Story = { args: { data: { ...list, items: [], pagination: { page: 1, pageSize: 25, total: 0 } } } };
export const CustomerEvent: Story = {
	args: { data: { ...list, items: [customerEvent, event], pagination: { page: 1, pageSize: 25, total: 2 } } },
};
export const NoMatches: Story = { args: { ...Empty.args, filtered: true } };
export const Pending: Story = { args: { data: undefined, loading: true } };
export const Error: Story = { args: { data: undefined, error: "Aktivitas tidak dapat dimuat." } };
export const RetainedError: Story = { args: { error: "Pembaruan gagal. Data sebelumnya masih ditampilkan." } };
export const LongLabels: Story = {
	args: {
		data: {
			...list,
			items: [
				{
					...event,
					actor: { ...event.actor, name: "Nama staf sangat panjang untuk memeriksa pembungkusan dan keterbacaan daftar operasional" },
					target: { ...event.target, name: "Organisasi operasional dengan identitas panjang yang tidak boleh membuat dokumen melebar" },
				},
			],
		},
	},
};
