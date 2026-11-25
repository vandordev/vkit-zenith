import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Button } from "@repo/ui/button";
import { AccessState } from "../../-components/access-state";
import { EditorDrawer, Notice } from "../../-components/common";
import { StaffPage } from "./-components/staff-page";
import { StaffGovernance } from "./-components/staff-governance";
import { usePrivateQuery, useProtected } from "../../../lib/api/session";
import * as sdk from "../../../lib/api/generated/sdk.gen";
import * as schemas from "../../../lib/api/generated/zod.gen";
import { errorMessage } from "../../../lib/api/client";
import { formInput, passwordValid } from "../../../lib/api/forms";
import { listSearch, useWrite } from "../../../lib/api/workspace";
import { staffListReturn } from "@bo/lib/api/customer-navigation";
export const Route = createFileRoute("/_authenticated/staff/")({
	validateSearch: listSearch.extend({ status: schemas.zBoV1ListStaffQuery.shape.status.catch("") }),
	// The former `?selected=` drawer is now the detail route; old links land there.
	beforeLoad: ({ search }) => {
		if (search.selected)
			throw redirect({
				to: "/staff/$staffId",
				params: { staffId: search.selected },
				search: { returnTo: staffListReturn(search) },
				replace: true,
			});
	},
	component: Staff,
});
function Staff() {
	const { staff } = useProtected();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const allowed = staff.role === "superadmin";
	const list = usePrivateQuery(
		["staff", search.q, search.status, search.page],
		async (t, signal) =>
			(
				await sdk.boV1ListStaff({
					client: t.client,
					signal,
					throwOnError: true,
					query: { q: search.q, status: search.status, page: search.page, pageSize: 25 },
				})
			).data,
		allowed,
	);
	if (!allowed) return <AccessState kind="denied" />;
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, create: undefined }), replace: true });
	};
	const create = () => {
		void navigate({ search: (previous) => ({ ...previous, create: true }) });
	};
	return (
		<StaffPage
			data={list.data}
			q={search.q}
			status={search.status ?? ""}
			navigationKey={JSON.stringify(search)}
			loading={list.isPending || list.isFetching}
			error={list.error ? errorMessage(list.error) : undefined}
			onFilter={(q, status) => {
				void navigate({
					search: (previous) => ({ ...previous, q, status: status === "active" || status === "disabled" ? status : "", page: 1 }),
				});
			}}
			onPage={(page) => {
				void navigate({ search: (previous) => ({ ...previous, page }) });
			}}
			onReload={() => {
				void list.refetch();
			}}
			onOpen={(id) => {
				if (!write.pending)
					void navigate({ to: "/staff/$staffId", params: { staffId: id }, search: { returnTo: staffListReturn(search) } });
			}}
			onCreate={create}
			renderLink={(item) => <Link to="/staff/$staffId" params={{ staffId: item.id }} search={{ returnTo: staffListReturn(search) }} />}
			action={
				<Button disabled={write.pending} onClick={create}>
					Tambah staf
				</Button>
			}
		>
			<Notice feedback={!search.create ? write.feedback : undefined} />
			<EditorDrawer open={!!search.create} editorKey="new" title="Tambah staf" pending={write.pending} onClose={close}>
				{(onDirty) =>
					search.create ? (
						<StaffGovernance
							key="new"
							creating
							pending={write.pending}
							feedback={write.feedback}
							onDirty={onDirty}
							onSubmit={(values) => {
								void write.run(async (t) => {
									if (!passwordValid(values.password ?? ""))
										throw new Error("Periksa kata sandi baru:15–128 karakter dan maksimal512 byte UTF-8.");
									await sdk.boV1CreateStaff({ ...(await t.writeOptions()), body: formInput(schemas.zBoV1CreateStaffBody, values) });
								}, close);
							}}
						/>
					) : null
				}
			</EditorDrawer>
		</StaffPage>
	);
}
