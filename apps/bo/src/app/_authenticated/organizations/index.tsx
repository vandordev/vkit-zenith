import { createFileRoute, Link } from "@tanstack/react-router";
import { OrganizationCollection } from "./-components/organization-collection";
import { OrganizationCreate } from "./-components/organization-create";
import { EditorDrawer, Notice } from "../../-components/common";
import { usePrivateQuery } from "@bo/lib/api/session";
import { useWrite } from "@bo/lib/api/workspace";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { organizationSearch, customerReturn } from "@bo/lib/api/customer-navigation";
import { organizationCreateBody } from "@bo/lib/api/customer-forms";
import { useCustomerChoices } from "./-adapters/use-customer-choices";
export const Route = createFileRoute("/_authenticated/organizations/")({ validateSearch: organizationSearch, component: Organizations });
function Organizations() {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const list = usePrivateQuery(
		["customer-organizations", search.q, search.status, search.page],
		async (t, signal) =>
			(
				await sdk.boV1ListOrganizations({
					client: t.client,
					signal,
					throwOnError: true,
					query: { q: search.q, status: search.status, page: search.page, pageSize: 25 },
				})
			).data,
	);
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, create: undefined }), replace: true });
	};
	return (
		<>
			<Notice feedback={!search.create ? write.feedback : undefined} />
			<OrganizationCollection
				data={list.data}
				q={search.q}
				status={search.status}
				navigationKey={JSON.stringify([search.status, search.page, search.create])}
				loading={list.isFetching}
				error={list.error ? errorMessage(list.error) : undefined}
				onCreate={() => {
					write.setFeedback(undefined);
					void navigate({ search: (previous) => ({ ...previous, create: true }) });
				}}
				onFilter={(q, status) => {
					void navigate({ search: (previous) => ({ ...previous, q, status, page: 1 }) });
				}}
				onPage={(page) => {
					void navigate({ search: (previous) => ({ ...previous, page }) });
				}}
				onRetry={() => {
					void list.refetch();
				}}
				organizationLink={(row) => (
					<Link
						to="/organizations/$organizationId"
						params={{ organizationId: row.id }}
						search={{ returnTo: customerReturn("/organizations", search) }}
					/>
				)}
			/>
			<EditorDrawer open={!!search.create} title="Tambah organisasi" onClose={close} pending={write.pending}>
				{(onDirty) =>
					search.create ? (
						<CreateAdapter
							pending={write.pending}
							feedback={write.feedback}
							onDirty={onDirty}
							onSubmit={(body) =>
								write.run(async (t) => {
									await sdk.boV1CreateOrganization({ ...(await t.writeOptions()), body: organizationCreateBody(body) });
								}, close)
							}
						/>
					) : null
				}
			</EditorDrawer>
		</>
	);
}
function CreateAdapter(props: Omit<React.ComponentProps<typeof OrganizationCreate>, "choices">) {
	const choices = useCustomerChoices(true);
	return <OrganizationCreate {...props} choices={choices} />;
}
