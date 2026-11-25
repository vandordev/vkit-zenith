import type { Meta, StoryObj } from "@storybook/react";
import { useLocation, useRouter } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { useState } from "react";
import { EditorDrawer } from "../../../-components/common";
import { organizationSearch } from "@bo/lib/api/customer-navigation";
import { OrganizationCollection } from "./organization-collection";
import { OrganizationCreate } from "./organization-create";
import { organization, organizations, ownerChoices, noop } from "../-fixtures";
const meta = {
	title: "BO/Pages/Organizations/Collection",
	component: OrganizationCollection,
	parameters: { layout: "fullscreen" },
	args: {
		data: organizations,
		q: "",
		status: "",
		onCreate: noop,
		onFilter: noop,
		onPage: noop,
		onRetry: noop,
		organizationLink: () => <a href="#organization" onClick={(event) => event.preventDefault()} />,
	},
	argTypes: {
		organizationLink: { control: false },
		onCreate: { control: false },
		onFilter: { control: false },
		onPage: { control: false },
		onRetry: { control: false },
	},
} satisfies Meta<typeof OrganizationCollection>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Populated: Story = {};
export const Interactive: Story = {
	render: function Preview(args) {
		const [q, setQ] = useState("");
		const [status, setStatus] = useState<"" | "active" | "suspended">("");
		const [message, setMessage] = useState("");
		const items = organizations.items.filter(
			(row) => (!q || `${row.name} ${row.slug}`.toLowerCase().includes(q.toLowerCase())) && (!status || row.status === status),
		);
		return (
			<>
				<OrganizationCollection
					{...args}
					q={q}
					status={status}
					data={{ ...organizations, items, pagination: { page: 1, pageSize: 25, total: String(items.length) } }}
					onFilter={(next, s) => {
						setQ(next);
						setStatus(s);
					}}
					onCreate={() => setMessage("Form organisasi diminta")}
					onRetry={() => setMessage("Muat ulang diminta")}
					organizationLink={(row) => (
						<a
							href={`#organization-${row.id}`}
							onClick={(event) => {
								event.preventDefault();
								setMessage(`Workspace ${row.name} diminta`);
							}}
						>
							{row.name}
						</a>
					)}
				/>
				{message && <p role="status">{message}</p>}
			</>
		);
	},
};
/** Actual collection + create drawer against Storybook memory history; pending search must not outlive create navigation. */
export const CreateNavigationHistory: Story = {
	parameters: { routerInitialPath: "/organizations?q=koperasi&page=3" },
	render: function Preview() {
		const location = useLocation();
		const router = useRouter();
		const search = organizationSearch.parse(location.search);
		const navigate = (create: boolean | undefined, replace = false) =>
			void router.navigate({ to: ".", replace, search: (previous) => ({ ...previous, create }) });
		return (
			<>
				<OrganizationCollection
					data={organizations}
					q={search.q}
					status={search.status}
					navigationKey={JSON.stringify([search.status, search.page, search.create])}
					onCreate={() => navigate(true)}
					onFilter={(q, status) => void router.navigate({ to: ".", search: (previous) => ({ ...previous, q, status, page: 1 }) })}
					onPage={(page) => void router.navigate({ to: ".", search: (previous) => ({ ...previous, page }) })}
					onRetry={noop}
					organizationLink={(row) => <a href={`#organization-${row.id}`}>{row.name}</a>}
				/>
				<div className="flex gap-2 px-4 pb-4">
					<Button type="button" variant="outline" onClick={() => router.history.forward()}>
						Riwayat maju
					</Button>
				</div>
				<p role="status" className="px-4">
					Diterapkan: q="{search.q}" halaman="{search.page}" buat="{search.create ? "terbuka" : "tertutup"}"
				</p>
				<EditorDrawer open={Boolean(search.create)} title="Tambah organisasi" onClose={() => navigate(undefined, true)}>
					{(onDirty) => (
						<div className="flex flex-col gap-3">
							<OrganizationCreate
								choices={{ data: ownerChoices, value: null, q: "", page: 1, onSelected: noop, onSearch: noop, onPage: noop, onRetry: noop }}
								onDirty={onDirty}
								onSubmit={noop}
							/>
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
export const Pending: Story = { args: { data: undefined, loading: true } };
export const Error: Story = { args: { data: undefined, error: "Organisasi tidak dapat dimuat." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan organisasi gagal." } };
export const Refreshing: Story = { args: { loading: true } };
export const Empty: Story = {
	args: { data: { items: [], pagination: { page: 1, pageSize: 25, total: "0" }, summary: { total: "0", active: "0", suspended: "0" } } },
};
export const NoMatches: Story = {
	args: { q: "Tidak ditemukan", data: { ...organizations, items: [], pagination: { page: 1, pageSize: 25, total: "0" } } },
};
export const LaterEmptyPage: Story = { args: { data: { ...organizations, items: [], pagination: { page: 3, pageSize: 25, total: "2" } } } };
export const ExactLargeCounts: Story = {
	args: {
		data: {
			...organizations,
			summary: { total: "9007199254740993", active: "9007199254740992", suspended: "1" },
			pagination: { page: 1, pageSize: 25, total: "9007199254740993" },
		},
	},
};
export const LongContent: Story = {
	args: {
		data: {
			...organizations,
			items: [
				{
					...organization,
					name: "Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara",
					slug: "koperasi-pengembangan-pendidikan-operasional-lintas-wilayah",
				},
			],
		},
	},
};
