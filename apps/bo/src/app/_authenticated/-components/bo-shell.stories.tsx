import type { Meta, StoryObj } from "@storybook/react";
import { useLocation } from "@tanstack/react-router";
import { PageContainer } from "@repo/ui/page-container";
import { BOShell } from "./bo-shell";
import { buildShellNavigation, buildSidebarContext } from "./navigation-links";
import { resolveNavigation, type NavContext } from "../-navigation";
import { parseCustomerReturn, parseStaffReturn } from "@bo/lib/api/customer-navigation";
import { noop, staff } from "../../-fixtures";

const org = "00000000-0000-4000-8000-000000000101";
// Uses the real resolver against the shared in-memory router: items, drill-down and the
// Live/Uji control all navigate, so the panel slide can be exercised in Storybook.
function Preview({ role, title, kind = "organization", children }: { role: "superadmin" | "admin"; title?: string; kind?: NavContext["kind"]; children?: string }) {
	const location = useLocation();
	const rawReturnTo = (location.search as { returnTo?: unknown }).returnTo;
	const returnTo = parseStaffReturn(rawReturnTo);
	const customerReturnTo = parseCustomerReturn(rawReturnTo);
	const context: NavContext | undefined = title ? { kind, id: org, title } : undefined;
	const navigation = resolveNavigation({ pathname: location.pathname, search: { returnTo, customerReturnTo }, role, context });
	return (
		<BOShell
			staff={{ ...staff, role }}
			navigation={buildShellNavigation(navigation.items)}
			navigationContext={navigation.panel}
			sidebarContext={buildSidebarContext(navigation.sidebarContext)}
			onLogout={noop}
			onAccount={noop}
			theme="system"
			onThemeChange={noop}
		>
			<PageContainer>
				<h1 className="text-2xl font-semibold">{children ?? "Konten halaman"}</h1>
				<p className="mt-2 text-sm text-muted-foreground">{location.pathname}</p>
			</PageContainer>
		</BOShell>
	);
}
const meta = {
	title: "BO/Layouts/Shell",
	component: BOShell,
	parameters: { layout: "fullscreen", routerInitialPath: "/" },
	render: () => <Preview role="superadmin" />,
	args: { staff, navigation: [], onLogout: noop, onAccount: noop, theme: "system", onThemeChange: noop, children: null },
	argTypes: { staff: { control: false }, navigation: { control: false }, children: { control: false } },
} satisfies Meta<typeof BOShell>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Global: Story = {};
export const Admin: Story = { render: () => <Preview role="admin" /> };
export const OrganizationContext: Story = {
	parameters: { routerInitialPath: `/organizations/${org}/api-keys` },
	render: () => <Preview role="admin" title="Toko Maju Jaya" />,
};
// Members context keeps q, status, page, selection and returnTo in the URL.
export const OrganizationContextKeepsSearch: Story = {
	parameters: {
		routerInitialPath: `/organizations/${org}/members?q=nadia&status=active&page=2&selected=${org}&returnTo=%2Forganizations%3Fpage%3D3`,
	},
	render: () => <Preview role="admin" title="Toko Maju Jaya" />,
};
export const OrganizationContextLongName: Story = {
	parameters: { routerInitialPath: `/organizations/${org}/members` },
	render: () => <Preview role="superadmin" title="Koperasi Pengembangan Pendidikan dan Operasional Pelanggan Lintas Wilayah Nusantara" />,
};
export const OrganizationContextLoading: Story = {
	parameters: { routerInitialPath: `/organizations/${org}` },
	render: () => <Preview role="admin" />,
};
export const GlobalOperationsAdmin: Story = {
	parameters: { routerInitialPath: "/operations/jobs" },
	render: () => <Preview role="admin" />,
};
/** User context: Ringkasan, Organisasi, Sesi login, Log Aktivitas. */
export const UserContext: Story = {
	parameters: { routerInitialPath: `/users/${org}/sessions` },
	render: () => <Preview role="admin" kind="user" title="Nadia Prameswari" />,
};
export const UserContextLongName: Story = {
	parameters: { routerInitialPath: `/users/${org}/organizations` },
	render: () => <Preview role="admin" kind="user" title="Nadia Prameswari Penanggung Jawab Operasional Organisasi Lintas Wilayah Nusantara" />,
};
export const UserContextKeepsMembershipOrigin: Story = {
	parameters: {
		routerInitialPath: `/users/${org}/activity?returnTo=${encodeURIComponent(`/organizations/${org}/members?q=nadia&status=active&ownership=owner&page=3&selected=${org}`)}`,
	},
	render: () => <Preview role="admin" kind="user" title="Nadia Prameswari" />,
};
export const UserContextLoading: Story = {
	parameters: { routerInitialPath: `/users/${org}` },
	render: () => <Preview role="admin" />,
};
/** Staff context is superadmin-only: Ringkasan, Sesi, Log Aktivitas. */
export const StaffContext: Story = {
	parameters: { routerInitialPath: `/staff/${org}/activity` },
	render: () => <Preview role="superadmin" kind="staff" title="Admin operasional seluruh cabang" />,
};
/** An admin never gets the staff panel; the global navigation stays. */
export const StaffContextDeniedForAdmin: Story = {
	parameters: { routerInitialPath: `/staff/${org}` },
	render: () => <Preview role="admin" kind="staff" title="Admin operasional seluruh cabang" />,
};
export const StaffContextKeepsListReturn: Story = {
	parameters: { routerInitialPath: `/staff/${org}/sessions?returnTo=${encodeURIComponent("/staff?q=Nadia&status=disabled&page=4")}` },
	render: () => <Preview role="superadmin" kind="staff" title="Nadia Prameswari" />,
};
