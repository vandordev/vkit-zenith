import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { QueryClient, QueryClientProvider, useQuery } from "@tanstack/react-query";
import { Button } from "@repo/ui/button";
import type { AuthTransport } from "../../../../lib/api/client";
import { keepSameScope, orgQueryKey } from "../../../../lib/organization-pure";

// Harness for the list observers' placeholder rule with a slow (1.5 s) fake query: the previous page stays only inside the same
// epoch and organization; another scope shows the pending state, never the previous partition's rows.
const slow = (label: string) =>
	new Promise<string[]>((resolve) => setTimeout(() => resolve([`${label} · baris 1`, `${label} · baris 2`]), 1500));

function Harness() {
	const [org, setOrg] = useState("org-a");
	const [page, setPage] = useState(1);
	const [epoch, setEpoch] = useState(1);
	const t = { epoch: () => epoch } as unknown as AuthTransport;
	const query = useQuery({
		queryKey: orgQueryKey(t, org, "api-keys", page),
		queryFn: () => slow(`${org}/halaman ${page}`),
		placeholderData: keepSameScope(t, org),
	});
	return (
		<div className="flex max-w-xl flex-col gap-4 p-6">
			<div className="flex flex-wrap gap-2">
				<Button type="button" variant="outline" onClick={() => setOrg(org === "org-a" ? "org-b" : "org-a")}>
					Ganti organisasi
				</Button>
				<Button type="button" variant="outline" onClick={() => setPage(page === 1 ? 2 : 1)}>
					Ganti halaman
				</Button>
				<Button type="button" variant="outline" onClick={() => setEpoch(epoch + 1)}>
					Identitas baru
				</Button>
			</div>
			<p className="text-sm text-muted-foreground">
				Sekarang: {org} · halaman {page} · epoch {epoch}
			</p>
			{query.data ? (
				<ul aria-label="Baris" className="rounded-md border border-border p-3 text-sm" data-placeholder={query.isPlaceholderData}>
					{query.data.map((row) => (
						<li key={row}>{row}</li>
					))}
					{query.isPlaceholderData && <li className="text-muted-foreground">Memuat halaman baru…</li>}
				</ul>
			) : (
				<p role="status" className="text-sm">
					Memuat…
				</p>
			)}
		</div>
	);
}

const meta = { title: "Web/Components/Dashboard/ListScopeHarness", component: Harness, parameters: { layout: "centered" } } satisfies Meta<
	typeof Harness
>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Same scope (page change) keeps the old rows with a loading hint; organization or identity changes show only "Memuat…". */
export const SlowQuery: Story = {
	render: () => <Harness key="harness" />,
	decorators: [
		(Story) => (
			<QueryClientProvider client={new QueryClient()}>
				<Story />
			</QueryClientProvider>
		),
	],
};
