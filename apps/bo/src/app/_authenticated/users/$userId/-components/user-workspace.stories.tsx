import type { Meta, StoryObj } from "@storybook/react";
import { Link, useLocation } from "@tanstack/react-router";
import { customerOrigin, userDetailSearch } from "@bo/lib/api/customer-navigation";
import { useState } from "react";
import { UserWorkspace } from "./user-workspace";
import { customer, verifiedCustomer } from "../../-fixtures";
import { emailActions, now } from "../-fixtures";
import { paged } from "../../../organizations/$organizationId/-setup-fixtures";
import { EmailActionList } from "./email-action-list";
import { noop } from "../../../organizations/-fixtures";
const meta = {
	title: "BO/Pages/Users/Workspace",
	component: UserWorkspace,
	parameters: { layout: "fullscreen" },
	args: {
		record: customer,
		emailActions: <EmailActionList actions={emailActions} state="ready" pagination={paged(emailActions.length)} now={now} onPage={noop} onRetry={noop} />,
		backLink: <a href="#users">Kembali ke pelanggan</a>,
		onRetry: noop,
		onEdit: noop,
		onEmail: noop,
		onVerify: noop,
		onStatus: noop,
	},
	render: function Preview(args) {
		const [message, setMessage] = useState("");
		return (
			<>
				<UserWorkspace
					{...args}
					onEdit={() => setMessage("Editor profil diminta")}
					onEmail={() => setMessage("Konfirmasi email diminta")}
					onVerify={() => setMessage("Konfirmasi verifikasi diminta")}
					onStatus={() => setMessage("Konfirmasi status diminta")}
				/>
				{message && <p role="status">{message}</p>}
			</>
		);
	},
} satisfies Meta<typeof UserWorkspace>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Admin: Story = {
	args: {
		record: {
			...customer,
			canChangeEmail: false,
			emailBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi.",
			canVerifyEmail: false,
			verificationBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi.",
			statusBlockedReason: "Hanya Superadmin dengan konfirmasi kata sandi.",
		},
	},
};
export const DisabledVerified: Story = { args: { record: verifiedCustomer } };
export const NoMemberships: Story = {
	args: {
		record: {
			...customer,
			membershipCount: "0",
			activeMembershipCount: "0",
			ownerMembershipCount: "0",
			canChangeStatus: true,
			statusBlockedReason: "",
		},
		emailActions: <EmailActionList actions={[]} state="ready" pagination={paged(0)} onPage={noop} onRetry={noop} />,
	},
};
export const EmailActionsPending: Story = {
	args: { emailActions: <EmailActionList actions={[]} state="pending" onPage={noop} onRetry={noop} /> },
};
export const EmailActionsError: Story = {
	args: { emailActions: <EmailActionList actions={[]} state="error" error="Tindakan email tidak dapat dimuat." onPage={noop} onRetry={noop} /> },
};
export const Pending: Story = { args: { record: undefined, loading: true, emailActions: undefined } };
export const Error: Story = { args: { record: undefined, error: "Pelanggan tidak tersedia." } };
export const RefreshFailure: Story = { args: { error: "Pembaruan pelanggan gagal." } };
export const LongEmail: Story = {
	args: {
		record: {
			...customer,
			email: "penanggung-jawab-operasional-pelanggan-lintas-wilayah@example.test",
			displayName: "Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah",
		},
	},
};
/** A validated activity return keeps its query and page through the workspace link in memory history. */
export const ActivityQueryReturn: Story = {
	parameters: {
		routerInitialPath: `/users/${customer.id}?returnTo=${encodeURIComponent(`/users/${customer.id}/activity?q=organization.updated&page=3`)}`,
	},
	render: function ActivityReturn(args) {
		const location = useLocation();
		const search = userDetailSearch.parse(location.search);
		const origin = new URL(customerOrigin(`/users/${customer.id}`, search), "https://bo.invalid");
		const target = { q: origin.searchParams.get("q") ?? "", page: Number(origin.searchParams.get("page") ?? 1) };
		return (
			<>
				<UserWorkspace
					{...args}
					backLink={
						<Link to="/users/$userId/activity" params={{ userId: customer.id }} search={target}>
							Kembali ke Log Aktivitas
						</Link>
					}
				/>
				<p role="status">Lokasi: {JSON.stringify({ pathname: location.pathname, search: location.search })}</p>
			</>
		);
	},
};
