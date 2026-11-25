import { createFileRoute, Link } from "@tanstack/react-router";
import { OrganizationWorkspace } from "./-components/organization-workspace";
import { OrganizationProfile } from "../-components/organization-profile";
import { OrganizationStatusAction } from "../-components/organization-status-action";
import { EditorDrawer, Notice } from "../../../-components/common";
import { usePrivateQuery } from "@bo/lib/api/session";
import { useWrite } from "@bo/lib/api/workspace";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { organizationOverviewSearch, parseCustomerReturn } from "@bo/lib/api/customer-navigation";
import * as forms from "@bo/lib/api/customer-forms";
export const Route = createFileRoute("/_authenticated/organizations/$organizationId/")({
	validateSearch: organizationOverviewSearch,
	component: Organization,
});
function Organization() {
	const { organizationId } = Route.useParams();
	return <OrganizationAdapter key={organizationId} organizationId={organizationId} />;
}
function OrganizationAdapter({ organizationId }: { organizationId: string }) {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const organization = usePrivateQuery(
		["customer-organization", organizationId],
		async (t, signal) =>
			(await sdk.boV1GetOrganization({ client: t.client, signal, throwOnError: true, path: { id: organizationId } })).data,
	);
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, action: undefined }), replace: true });
	};
	const open = (action: "profile" | "status") => {
		write.setFeedback(undefined);
		void navigate({ search: (previous) => ({ ...previous, action }) });
	};
	const reloadOrg = async () => {
		const fresh = await organization.refetch();
		return fresh.error ? undefined : fresh.data;
	};
	const record = organization.data;
	return (
		<>
			<Notice feedback={!search.action ? write.feedback : undefined} />
			<OrganizationWorkspace
				record={record}
				loading={organization.isFetching}
				error={organization.error ? errorMessage(organization.error) : undefined}
				backLink={
					<Link className="text-sm underline" to={parseCustomerReturn(search.returnTo) ?? "/organizations"}>
						Kembali ke daftar asal
					</Link>
				}
				onRetry={() => {
					void organization.refetch();
				}}
				onEdit={() => open("profile")}
				onStatus={() => open("status")}
				onOwners={() => {
					void navigate({
						to: "/organizations/$organizationId/members",
						params: { organizationId },
						search: { q: "", status: "", ownership: "owner", page: 1 },
					});
				}}
			/>
			<EditorDrawer
				open={!!record && !!search.action}
				title={search.action === "status" ? "Ubah status organisasi" : "Ubah profil organisasi"}
				editorKey={`${organizationId}:${search.action ?? ""}`}
				pending={write.pending}
				onClose={close}
			>
				{(onDirty) =>
					record ? (
						search.action === "status" ? (
							<OrganizationStatusAction
								record={record}
								pending={write.pending}
								feedback={write.feedback}
								onReload={reloadOrg}
								onDirty={onDirty}
								onSubmit={(body) =>
									write.run(async (t) => {
										await sdk.boV1SetOrganizationStatus({
											...(await t.writeOptions()),
											path: { id: organizationId },
											body: forms.organizationStatusBody(body),
										});
									}, close)
								}
							/>
						) : (
							<OrganizationProfile
								record={record}
								pending={write.pending}
								feedback={write.feedback}
								refreshError={organization.error ? errorMessage(organization.error) : undefined}
								onReload={reloadOrg}
								onDirty={onDirty}
								onSubmit={(body) =>
									write.run(async (t) => {
										await sdk.boV1UpdateOrganization({
											...(await t.writeOptions()),
											path: { id: organizationId },
											body: forms.organizationProfileBody(body),
										});
									}, close)
								}
							/>
						)
					) : null
				}
			</EditorDrawer>
		</>
	);
}
