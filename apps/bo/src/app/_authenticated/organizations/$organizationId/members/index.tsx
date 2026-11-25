import { createFileRoute, Link } from "@tanstack/react-router";
import { MembersPage } from "./-components/members-page";
import { MembershipCollection } from "./-components/membership-collection";
import { MembershipAdd, MembershipEditor } from "./-components/membership-editor";
import { OwnerConfirmation } from "./-components/owner-confirmation";
import { EditorDrawer, Notice } from "../../../../-components/common";
import { ContentState } from "../../../../-components/content-state";
import { usePrivateQuery } from "@bo/lib/api/session";
import { useWrite } from "@bo/lib/api/workspace";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { organizationMembersSearch, customerOrigin, parseCustomerReturn } from "@bo/lib/api/customer-navigation";
import * as forms from "@bo/lib/api/customer-forms";
import { useCustomerChoices } from "../../-adapters/use-customer-choices";
export const Route = createFileRoute("/_authenticated/organizations/$organizationId/members/")({
	validateSearch: organizationMembersSearch,
	component: Members,
});
function Members() {
	const { organizationId } = Route.useParams();
	return <MembersAdapter key={organizationId} organizationId={organizationId} />;
}
function MembersAdapter({ organizationId }: { organizationId: string }) {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const organization = usePrivateQuery(
		["customer-organization", organizationId],
		async (t, signal) =>
			(await sdk.boV1GetOrganization({ client: t.client, signal, throwOnError: true, path: { id: organizationId } })).data,
	);
	const members = usePrivateQuery(
		["customer-memberships", organizationId, search.q, search.status, search.ownership, search.page],
		async (t, signal) =>
			(
				await sdk.boV1ListOrganizationMemberships({
					client: t.client,
					signal,
					throwOnError: true,
					path: { organizationId },
					query: { q: search.q, status: search.status, ownership: search.ownership, page: search.page, pageSize: 25 },
				})
			).data,
		!!organization.data,
	);
	const detail = usePrivateQuery(
		["customer-membership", organizationId, search.selected],
		async (t, signal) =>
			(
				await sdk.boV1GetOrganizationMembership({
					client: t.client,
					signal,
					throwOnError: true,
					path: { organizationId, id: search.selected ?? "" },
				})
			).data,
		!!search.selected && !!organization.data,
	);
	const member = detail.data?.organizationId === organizationId && detail.data.id === search.selected ? detail.data : undefined;
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, selected: undefined, action: undefined }), replace: true });
	};
	const open = (action: "add") => {
		write.setFeedback(undefined);
		void navigate({ search: (previous) => ({ ...previous, action, selected: undefined }) });
	};
	const reloadMember = async () => {
		const fresh = await detail.refetch();
		return fresh.error ? undefined : fresh.data;
	};
	const origin = customerOrigin(`/organizations/${organizationId}/members`, search);
	const userLink = (row: { userId: string; userName: string }) => (
		<Link
			className="hover:underline focus-visible:underline"
			to="/users/$userId"
			params={{ userId: row.userId }}
			search={{ page: 1, returnTo: origin }}
		>
			{row.userName}
		</Link>
	);
	const title = search.action === "add" ? "Tambah anggota" : search.action === "owner" ? "Konfirmasi kepemilikan" : "Keanggotaan pelanggan";
	const record = organization.data;
	return (
		<>
			<Notice feedback={!search.action && !search.selected ? write.feedback : undefined} />
			<MembersPage
				error={organization.error ? errorMessage(organization.error) : undefined}
				backLink={
					<Link className="text-sm underline" to={parseCustomerReturn(search.returnTo) ?? "/organizations"}>
						Kembali ke daftar asal
					</Link>
				}
				onRetry={() => {
					void organization.refetch();
				}}
			>
				<MembershipCollection
					data={members.data}
					filters={{ q: search.q, status: search.status, ownership: search.ownership }}
					navigationKey={JSON.stringify([organizationId, search.status, search.ownership, search.page, search.selected, search.action, search.returnTo])}
					loading={members.isFetching}
					error={members.error ? errorMessage(members.error) : undefined}
					onFilter={(filters) => {
						void navigate({ search: (previous) => ({ ...previous, ...filters, page: 1 }) });
					}}
					onPage={(page) => {
						void navigate({ search: (previous) => ({ ...previous, page, selected: undefined, action: undefined }) });
					}}
					onRetry={() => {
						void members.refetch();
					}}
					onAdd={() => open("add")}
					userLink={userLink}
					onSelect={(row) => {
						write.setFeedback(undefined);
						void navigate({ search: (previous) => ({ ...previous, selected: row.id, action: undefined }) });
					}}
				/>
			</MembersPage>
			<EditorDrawer
				open={!!record && !!(search.action || search.selected)}
				title={title}
				editorKey={`${organizationId}:${search.action ?? "member"}:${search.selected ?? ""}`}
				pending={write.pending}
				onClose={close}
			>
				{(onDirty) =>
					record ? (
						<>
							{search.action === "add" ? (
								<AddAdapter
									organizationId={organizationId}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1AddOrganizationMember({
												...(await t.writeOptions()),
												path: { organizationId },
												body: forms.memberAddBody(body),
											});
										}, close)
									}
								/>
							) : search.action === "owner" && member ? (
								<OwnerConfirmation
									record={member}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onReload={reloadMember}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1SetOrganizationMembershipRole({
												...(await t.writeOptions()),
												path: { organizationId, id: member.id },
												body: forms.ownerBody(body),
											});
										}, close)
									}
								/>
							) : search.action === "owner" ? (
								<ContentState
									{...(detail.error
										? {
												kind: "error",
												title: errorMessage(detail.error),
												onRetry: () => {
													void detail.refetch();
												},
											}
										: { kind: "loading", title: "Memuat keanggotaan…" })}
								/>
							) : (
								<MembershipEditor
									record={member}
									loading={detail.isFetching}
									error={detail.error ? errorMessage(detail.error) : undefined}
									feedback={write.feedback}
									pending={write.pending}
									userLink={member ? userLink(member) : undefined}
									onRetry={() => {
										void detail.refetch();
									}}
									onOwner={() => {
										write.setFeedback(undefined);
										void navigate({ search: (previous) => ({ ...previous, action: "owner" }), replace: true });
									}}
									onRemove={(body) =>
										member &&
										write.run(async (t) => {
											await sdk.boV1RemoveOrganizationMember({
												...(await t.writeOptions()),
												path: { organizationId, id: member.id },
												body: forms.memberRemoveBody(body),
											});
										}, close)
									}
									onRestore={(body) =>
										member &&
										write.run(async (t) => {
											await sdk.boV1RestoreOrganizationMember({
												...(await t.writeOptions()),
												path: { organizationId, id: member.id },
												body: forms.memberRestoreBody(body),
											});
										}, close)
									}
								/>
							)}
						</>
					) : null
				}
			</EditorDrawer>
		</>
	);
}
function AddAdapter({
	organizationId,
	...props
}: Omit<React.ComponentProps<typeof MembershipAdd>, "choices"> & { organizationId: string }) {
	const choices = useCustomerChoices(true, organizationId);
	return <MembershipAdd {...props} choices={choices} />;
}
