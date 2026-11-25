import { createFileRoute, Link } from "@tanstack/react-router";
import { UserWorkspace } from "./-components/user-workspace";
import { EmailActionList } from "./-components/email-action-list";
import { UserProfile } from "./-components/user-profile";
import { UserEmailAction } from "./-components/user-email-action";
import { UserVerificationAction } from "./-components/user-verification-action";
import { UserStatusAction } from "./-components/user-status-action";
import { EditorDrawer, Notice } from "../../../-components/common";
import { usePrivateQuery } from "@bo/lib/api/session";
import { useWrite } from "@bo/lib/api/workspace";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import * as forms from "@bo/lib/api/customer-forms";
import { userDetailSearch, parseCustomerReturn } from "@bo/lib/api/customer-navigation";
import { pageSize, sectionOf } from "../../organizations/$organizationId/-adapters/org-adapter";
export const Route = createFileRoute("/_authenticated/users/$userId/")({ validateSearch: userDetailSearch, component: User });
function User() {
	const { userId } = Route.useParams();
	return <UserAdapter key={userId} userId={userId} />;
}
function UserAdapter({ userId }: { userId: string }) {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const write = useWrite();
	const user = usePrivateQuery(
		["customer-user", userId],
		async (t, signal) => (await sdk.boV1GetUser({ client: t.client, signal, throwOnError: true, path: { id: userId } })).data,
	);
	// Email actions are the only paged list left on the overview; memberships live under Organisasi.
	const emailActions = usePrivateQuery(
		["customer-user-email-actions", userId, search.page],
		async (t, signal) =>
			(await sdk.boV1ListUserEmailActions({ client: t.client, signal, throwOnError: true, path: { userId }, query: { page: search.page, pageSize } })).data,
		!!user.data,
	);
	const close = () => {
		void navigate({ search: (previous) => ({ ...previous, action: undefined }), replace: true });
	};
	const open = (action: "profile" | "email" | "verify" | "status") => {
		write.setFeedback(undefined);
		void navigate({ search: (previous) => ({ ...previous, action }) });
	};
	const reload = async () => {
		const fresh = await user.refetch();
		return fresh.error ? undefined : fresh.data;
	};
	const record = user.data;
	return (
		<>
			<Notice feedback={!search.action ? write.feedback : undefined} />
			<UserWorkspace
				record={record}
				loading={user.isFetching}
				error={user.error ? errorMessage(user.error) : undefined}
				emailActions={
					<EmailActionList
						actions={emailActions.data?.items ?? []}
						pagination={emailActions.data?.pagination}
						loading={emailActions.isFetching}
						{...sectionOf(emailActions)}
						onPage={(page) => {
							void navigate({ search: (previous) => ({ ...previous, page }) });
						}}
						onRetry={() => {
							void emailActions.refetch();
						}}
					/>
				}
				backLink={
					<Link className="text-sm underline" to={parseCustomerReturn(search.returnTo) ?? "/users"}>
						Kembali ke daftar asal
					</Link>
				}
				onRetry={() => {
					void user.refetch();
				}}
				onEdit={() => open("profile")}
				onEmail={() => open("email")}
				onVerify={() => open("verify")}
				onStatus={() => open("status")}
			/>
			<EditorDrawer
				open={!!record && !!search.action}
				title={
					search.action === "profile"
						? "Ubah profil pelanggan"
						: search.action === "email"
							? "Ganti email"
							: search.action === "verify"
								? "Verifikasi email manual"
								: "Ubah status pelanggan"
				}
				editorKey={`${userId}:${search.action ?? ""}`}
				pending={write.pending}
				onClose={close}
			>
				{(onDirty) =>
					record ? (
						<>
							{search.action === "profile" ? (
								<UserProfile
									record={record}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onReload={reload}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1UpdateUser({ ...(await t.writeOptions()), path: { id: userId }, body: forms.userProfileBody(body) });
										}, close)
									}
								/>
							) : search.action === "email" ? (
								<UserEmailAction
									record={record}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onReload={reload}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1ChangeUserEmail({ ...(await t.writeOptions()), path: { id: userId }, body: forms.userEmailBody(body) });
										}, close)
									}
								/>
							) : search.action === "verify" ? (
								<UserVerificationAction
									record={record}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onReload={reload}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1VerifyUserEmail({
												...(await t.writeOptions()),
												path: { id: userId },
												body: forms.userVerifyBody(body),
											});
										}, close)
									}
								/>
							) : search.action === "status" ? (
								<UserStatusAction
									record={record}
									pending={write.pending}
									feedback={write.feedback}
									onDirty={onDirty}
									onReload={reload}
									onSubmit={(body) =>
										write.run(async (t) => {
											await sdk.boV1SetUserStatus({ ...(await t.writeOptions()), path: { id: userId }, body: forms.userStatusBody(body) });
										}, close)
									}
								/>
							) : null}
						</>
					) : null
				}
			</EditorDrawer>
		</>
	);
}
