import { createFileRoute, Link } from "@tanstack/react-router";
import { UserCollection } from "./-components/user-collection";
import { UserCreate } from "./-components/user-create";
import { EditorDrawer, Notice } from "../../-components/common";
import { usePrivateQuery } from "@bo/lib/api/session";
import { useWrite } from "@bo/lib/api/workspace";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { userSearch, customerReturn } from "@bo/lib/api/customer-navigation";
import { userCreateBody } from "@bo/lib/api/customer-forms";
export const Route = createFileRoute("/_authenticated/users/")({ validateSearch: userSearch, component: Users });
function Users() {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const list = usePrivateQuery(
		["customer-users", search.q, search.status, search.verification, search.page],
		async (t, signal) =>
			(
				await sdk.boV1ListUsers({
					client: t.client,
					signal,
					throwOnError: true,
					query: { q: search.q, status: search.status, verification: search.verification, page: search.page, pageSize: 25 },
				})
			).data,
	);
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, create: undefined }), replace: true });
	};
	return (
		<>
			<Notice feedback={!search.create ? write.feedback : undefined} />
			<UserCollection
				data={list.data}
				filters={{ q: search.q, status: search.status, verification: search.verification }}
				navigationKey={JSON.stringify([search.status, search.verification, search.page, search.create])}
				loading={list.isFetching}
				error={list.error ? errorMessage(list.error) : undefined}
				onCreate={() => {
					write.setFeedback(undefined);
					void navigate({ search: (previous) => ({ ...previous, create: true }) });
				}}
				onFilter={(filters) => {
					void navigate({ search: (previous) => ({ ...previous, ...filters, page: 1 }) });
				}}
				onPage={(page) => {
					void navigate({ search: (previous) => ({ ...previous, page }) });
				}}
				onRetry={() => {
					void list.refetch();
				}}
				userLink={(row) => (
					<Link to="/users/$userId" params={{ userId: row.id }} search={{ page: 1, returnTo: customerReturn("/users", search) }} />
				)}
			/>
			<EditorDrawer open={!!search.create} title="Tambah pelanggan" pending={write.pending} onClose={close}>
				{(onDirty) =>
					search.create ? (
						<UserCreate
							pending={write.pending}
							feedback={write.feedback}
							onDirty={onDirty}
							onSubmit={(body) =>
								write.run(async (t) => {
									await sdk.boV1CreateUser({ ...(await t.writeOptions()), body: userCreateBody(body) });
								}, close)
							}
						/>
					) : null
				}
			</EditorDrawer>
		</>
	);
}
