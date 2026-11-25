import type { Meta, StoryObj } from "@storybook/react";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { OrgOverview } from "./org-overview";
import { OrgShell } from "./org-shell";
import { linkProps } from "./navigation";
import { longOrganization, navigationFor, organization, organizations } from "./-fixtures";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";

const noop = () => {};
const slug = "kopi-senja";

// Uses the real resolver against the shared in-memory router: items and the Pengaturan drill-down
// navigate, so the panel slide can be exercised here.
function Preview({
	org = organization(),
	list = [org],
	defaultCollapsed,
}: {
	org?: V1MyOrganization;
	list?: V1MyOrganization[];
	defaultCollapsed?: boolean;
}) {
	const location = useLocation();
	const navigate = useNavigate();
	return (
		<OrgShell
			organization={org}
			organizations={list}
			defaultCollapsed={defaultCollapsed}
			navigation={navigationFor(org.slug, location.pathname)}
			userName="Rani Wulandari"
			onLogout={noop}
			theme="system"
			onThemeChange={noop}
			onNavigate={(link) => void navigate(linkProps(link))}
		>
			<OrgOverview organization={org} memberCount={4} membersState="ready" onRetryMembers={noop} />
		</OrgShell>
	);
}

const meta = {
	title: "Web/Layouts/OrgShell",
	component: OrgShell,
	parameters: { layout: "fullscreen", routerInitialPath: `/w/${slug}` },
	render: () => <Preview />,
	args: {
		organization: organization(),
		organizations: [organization()],
		navigation: navigationFor(slug),
		userName: "Rani Wulandari",
		onLogout: noop,
		theme: "system",
		onThemeChange: noop,
		onNavigate: noop,
		children: null,
	},
	argTypes: {
		organization: { control: false },
		organizations: { control: false },
		navigation: { control: false },
		children: { control: false },
	},
} satisfies Meta<typeof OrgShell>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Owner: Story = {};
export const Admin: Story = { render: () => <Preview org={organization("admin")} /> };
export const Member: Story = { render: () => <Preview org={organization("member")} /> };
export const Suspended: Story = { render: () => <Preview org={organization("owner", { status: "suspended" })} /> };
export const SettingsPanel: Story = { parameters: { routerInitialPath: `/w/${slug}/settings/profile` } };
export const MultipleOrganizations: Story = { render: () => <Preview list={organizations} /> };
export const Collapsed: Story = { render: () => <Preview defaultCollapsed list={organizations} /> };
export const Mobile: Story = { globals: { viewport: { value: "mobile1", isRotated: false } } };
export const LongOrganizationName: Story = {
	render: () => <Preview list={[longOrganization, ...organizations.slice(1)]} org={longOrganization} />,
};
