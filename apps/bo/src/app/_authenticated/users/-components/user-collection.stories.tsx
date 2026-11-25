import type { Meta, StoryObj } from "@storybook/react";
import { useLocation, useRouter } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { EditorDrawer } from "../../../-components/common";
import { userSearch } from "@bo/lib/api/customer-navigation";
import { UserCollection } from "./user-collection";
import { UserCreate } from "./user-create";
import { customer, customers } from "../-fixtures";
import { noop } from "../../organizations/-fixtures";
const meta = {
	title: "BO/Pages/Users/Collection",
	component: UserCollection,
	parameters: { layout: "fullscreen" },
	args: {
		data: customers,
		filters: { q: "", status: "", verification: "" },
		onCreate: noop,
		onFilter: noop,
		onPage: noop,
		onRetry: noop,
		userLink: (row) => <a href={`#user-${row.id}`} />,
	},
} satisfies Meta<typeof UserCollection>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Empty: Story = {
	args: {
		data: {
			items: [],
			pagination: { page: 1, pageSize: 25, total: "0" },
			summary: { total: "0", active: "0", disabled: "0", verified: "0", unverified: "0" },
		},
	},
};
export const NoMatches: Story = {
	args: {
		filters: { q: "Tidak ada", status: "disabled", verification: "unverified" },
		data: { ...customers, items: [], pagination: { page: 1, pageSize: 25, total: "0" } },
	},
};
export const Pending: Story = { args: { data: undefined, loading: true } };
export const Error: Story = { args: { data: undefined, error: "Pelanggan tidak dapat dimuat." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan pelanggan gagal." } };
export const LongContent: Story = {
	args: {
		data: {
			...customers,
			items: [
				{
					...customer,
					displayName: "Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah",
					email: "penanggung-jawab-operasional-pelanggan-lintas-wilayah@example.test",
				},
			],
		},
	},
};
/** Actual collection + create drawer against Storybook memory history; pending search must not outlive create navigation. */
export const CreateNavigationHistory: Story = {
	parameters: { routerInitialPath: "/users?q=nadia&page=3" },
	render: function Preview() {
		const location = useLocation();
		const router = useRouter();
		const search = userSearch.parse(location.search);
		const navigate = (create: boolean | undefined, replace = false) =>
			void router.navigate({ to: ".", replace, search: (previous) => ({ ...previous, create }) });
		return (
			<>
				<UserCollection
					data={customers}
					filters={{ q: search.q, status: search.status, verification: search.verification }}
					navigationKey={JSON.stringify([search.status, search.verification, search.page, search.create])}
					onCreate={() => navigate(true)}
					onFilter={(filters) => void router.navigate({ to: ".", search: (previous) => ({ ...previous, ...filters, page: 1 }) })}
					onPage={(page) => void router.navigate({ to: ".", search: (previous) => ({ ...previous, page }) })}
					onRetry={noop}
					userLink={(row) => <a href={`#user-${row.id}`} />}
				/>
				<div className="flex gap-2 px-4 pb-4">
					<Button type="button" variant="outline" onClick={() => router.history.forward()}>
						Riwayat maju
					</Button>
				</div>
				<p role="status" className="px-4">
					Diterapkan: q="{search.q}" halaman="{search.page}" buat="{search.create ? "terbuka" : "tertutup"}"
				</p>
				<EditorDrawer open={Boolean(search.create)} title="Tambah pelanggan" onClose={() => navigate(undefined, true)}>
					{(onDirty) => (
						<div className="flex flex-col gap-3">
							<UserCreate onDirty={onDirty} onSubmit={noop} />
							<Button type="button" variant="outline" onClick={() => router.history.back()}>
								Riwayat kembali
							</Button>
						</div>
					)}
				</EditorDrawer>
			</>
		);
	},
};
export const LargeCounts: Story = {
	args: {
		data: {
			...customers,
			summary: { total: "9007199254740993", active: "9007199254740992", disabled: "1", verified: "1", unverified: "9007199254740992" },
		},
	},
};
