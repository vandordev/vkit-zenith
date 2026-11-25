import type { Meta, StoryObj } from "@storybook/react";
import { ArrowLeftIcon, BotIcon, FolderIcon, LayoutGridIcon, SettingsIcon } from "lucide-react";
import { useState } from "react";

import { AppShell } from "./app-shell";
import { Button } from "./button";
import { PageContainer } from "./page-container";

const meta = {
	argTypes: {
		actions: { control: false },
		brand: { control: false },
		children: { control: false },
		collapsed: { control: "boolean" },
		defaultCollapsed: { control: false },
		header: { control: false },
		showHeader: { control: "boolean" },
		inset: { control: "boolean" },
		navigation: { control: false },
		navigationContext: { control: false },
		navigationLabel: { control: "text" },
		onCollapsedChange: { control: false },
		onSearch: { control: false },
		searchLabel: { control: "text" },
		sidebarFooter: { control: false },
		sidebarHeader: { control: false },
	},
	args: {
		showHeader: true,
		brand: { icon: <LayoutGridIcon />, name: "Acme" },
		children: (
			<PageContainer>
				<h1 className="text-2xl font-semibold">Workspace</h1>
				<p className="mt-2 text-sm text-muted-foreground">Consumer content belongs here.</p>
			</PageContainer>
		),
		navigation: [
			{
				active: true,
				icon: <LayoutGridIcon />,
				id: "overview",
				label: "Overview",
			},
			{ icon: <FolderIcon />, id: "projects", label: "Projects" },
			{ icon: <SettingsIcon />, id: "settings", label: "Settings" },
		],
	},
	component: AppShell,
	parameters: { layout: "fullscreen" },
	title: "UI/AppShell",
} satisfies Meta<typeof AppShell>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Collapsed: Story = { args: { collapsed: true } };
export const Headerless: Story = {
	args: { showHeader: false },
	parameters: {
		docs: {
			description: {
				story:
					"Desktop controls live in the sidebar. On mobile, a top-attached brand tab with inverted upper corners and a floating toggle stay fixed at opposite corners, without a header. The toggle opens a right-side drawer.",
			},
		},
	},
};
export const HeaderlessCollapsed: Story = { args: { showHeader: false, defaultCollapsed: true } };
export const Minimal: Story = { args: { navigation: [] } };
export const FullBleed: Story = {
	args: {
		children: <div className="min-h-64 bg-muted text-sm">Edge-to-edge content without a PageContainer.</div>,
	},
};
export const DisabledDestination: Story = {
	args: {
		navigation: [
			{
				disabled: true,
				icon: <SettingsIcon />,
				id: "locked",
				label: "Unavailable settings",
			},
		],
	},
};
export const Controlled: Story = {
	argTypes: { collapsed: { control: false } },
	render: function ControlledShell(args) {
		const [collapsed, setCollapsed] = useState(false);
		return (
			<AppShell
				{...args}
				collapsed={collapsed}
				onCollapsedChange={setCollapsed}
				header={<span className="text-sm">{collapsed ? "Collapsed" : "Expanded"}</span>}
			/>
		);
	},
};
export const NavigationAndSearch: Story = {
	render: function InteractiveShell(args) {
		const [active, setActive] = useState("overview");
		const [status, setStatus] = useState("Select a destination or use Cmd/Ctrl+K.");
		return (
			<AppShell
				{...args}
				navigation={args.navigation.map((item) => ({
					...item,
					active: item.id === active,
					onClick: () => {
						setActive(item.id);
						setStatus(`${item.label} selected`);
					},
				}))}
				onSearch={() => setStatus("Search requested. The application owns its search UI.")}
			>
				<PageContainer>
					<h1 className="text-2xl font-semibold">{active}</h1>
					<p role="status" className="mt-2 text-sm text-muted-foreground">
						{status}
					</p>
				</PageContainer>
			</AppShell>
		);
	},
};
export const WorkspaceAndAccount: Story = {
	render: function WorkspaceShell(args) {
		const [workspace, setWorkspace] = useState("Acme");
		const [notice, setNotice] = useState("No account action selected");
		return (
			<AppShell
				{...args}
				sidebarHeader={({ collapsed, closeNavigation }) => (
					<Button
						variant="ghost"
						className="h-12 w-full justify-start overflow-hidden px-3"
						aria-label={`Switch workspace: ${workspace}`}
						onClick={() => {
							setWorkspace(workspace === "Acme" ? "Orbit" : "Acme");
							closeNavigation();
						}}
					>
						<LayoutGridIcon aria-hidden="true" />
						{!collapsed && <span className="truncate">{workspace}</span>}
					</Button>
				)}
				sidebarFooter={({ collapsed }) => (
					<Button
						variant="ghost"
						className="h-12 w-full justify-start overflow-hidden px-3"
						aria-label="Account action"
						onClick={() => setNotice("Local account action selected")}
					>
						<SettingsIcon aria-hidden="true" />
						{!collapsed && <span className="truncate">Avery Cole</span>}
					</Button>
				)}
			>
				<PageContainer>
					<h1 className="text-2xl font-semibold">{workspace}</h1>
					<p role="status" className="mt-2 text-sm text-muted-foreground">
						{notice}
					</p>
				</PageContainer>
			</AppShell>
		);
	},
};
export const Hierarchy: Story = {
	render: function HierarchyShell(args) {
		const [context, setContext] = useState<"workspace" | "agent">("workspace");
		const [agent, setAgent] = useState("Customer Support");
		const [section, setSection] = useState("Profile");
		return (
			<AppShell
				{...args}
				navigationContext={{
					depth: context === "workspace" ? 0 : 1,
					key: context,
				}}
				navigation={
					context === "workspace"
						? [
								{
									active: true,
									icon: <BotIcon />,
									id: "agents",
									label: "Agents",
									onClick: () => setContext("agent"),
								},
							]
						: [
								{
									dividerAfter: true,
									icon: <ArrowLeftIcon />,
									id: "back",
									label: "All agents",
									onClick: () => setContext("workspace"),
								},
								{
									id: "switcher",
									label: agent,
									renderControl: ({ collapsed, closeNavigation }) => (
										<Button
											variant="ghost"
											className="h-12 w-full justify-start overflow-hidden px-3"
											aria-label={`Switch agent: ${agent}`}
											onClick={() => {
												setAgent(agent === "Customer Support" ? "Research Assistant" : "Customer Support");
												closeNavigation();
											}}
										>
											<BotIcon aria-hidden="true" />
											{!collapsed && <span className="truncate">{agent}</span>}
										</Button>
									),
								},
								...["Profile", "Configure", "Versions"].map((label) => ({
									active: section === label,
									id: label,
									label,
									onClick: () => setSection(label),
								})),
							]
				}
			>
				<PageContainer>
					<h1 className="text-2xl font-semibold">{context === "workspace" ? "Agents" : `${agent}: ${section}`}</h1>
					{context === "workspace" && (
						<Button className="mt-4" onClick={() => setContext("agent")}>
							Open Customer Support
						</Button>
					)}
				</PageContainer>
			</AppShell>
		);
	},
};
export const LongContent: Story = {
	args: {
		brand: { name: "Acme Research and Development Workspace" },
		children: (
			<PageContainer>
				{Array.from({ length: 40 }, (_, index) => (
					<p key={index} className="py-3 text-sm">
						Workspace activity {index + 1}
					</p>
				))}
			</PageContainer>
		),
		navigation: Array.from({ length: 35 }, (_, index) => ({
			active: index === 0,
			icon: <FolderIcon />,
			id: `project-${index}`,
			label: `Research project ${index + 1}: a deliberately long navigation label`,
		})),
	},
};
export const Mobile: Story = {
	parameters: {
		docs: {
			description: {
				story:
					"Use a viewport below 768px. Open navigation, select a destination, and verify dismissal/focus return. Desktop collapsed state is independent from the drawer.",
			},
		},
	},
};
export const ReducedMotion: Story = {
	parameters: {
		docs: {
			description: {
				story:
					"Enable prefers-reduced-motion in your browser or OS. Collapse, labels and hierarchy transitions complete without travel animations; no addon is required.",
			},
		},
	},
	render: Hierarchy.render,
};
export const HeaderlessControlled: Story = { ...Controlled, args: { showHeader: false } };
export const HeaderlessCustomHeader: Story = { ...WorkspaceAndAccount, args: { showHeader: false } };
export const HeaderlessLongContent: Story = { ...LongContent, args: { ...LongContent.args, showHeader: false } };

const contextItems = [
	{ active: true, icon: <FolderIcon />, id: "overview", label: "Ringkasan" },
	{ icon: <SettingsIcon />, id: "settings", label: "Pengaturan" },
];
const ContextShell = ({ title, ...args }: React.ComponentProps<typeof AppShell> & { title: string }) => {
	const [inside, setInside] = useState(true);
	return (
		<AppShell
			{...args}
			navigation={
				inside ? contextItems : [{ active: true, icon: <BotIcon />, id: "orgs", label: "Organisasi", onClick: () => setInside(true) }]
			}
			navigationContext={{ depth: inside ? 1 : 0, key: inside ? "organization" : "global" }}
			sidebarContext={
				inside
					? {
							back: {
								label: "Semua organisasi",
								render: ({ collapsed }) => (
									<Button
										variant="ghost"
										size={collapsed ? "icon" : "sm"}
										aria-label="Semua organisasi"
										className={collapsed ? undefined : "-ml-2"}
										onClick={() => setInside(false)}
									>
										<ArrowLeftIcon aria-hidden="true" />
										{!collapsed && "Semua organisasi"}
									</Button>
								),
							},
							title,
						}
					: undefined
			}
		>
			<PageContainer>
				<p className="mt-4 text-sm text-muted-foreground">Konten organisasi.</p>
			</PageContainer>
		</AppShell>
	);
};
export const ContextPanel: Story = {
	render: (args) => <ContextShell {...args} title="Toko Maju Jaya" />,
	parameters: {
		docs: { description: { story: "Back control slides to the global panel (reverse); selecting Organisasi enters again (forward)." } },
	},
};
export const ContextPanelCollapsed: Story = {
	args: { collapsed: true },
	render: (args) => <ContextShell {...args} title="Toko Maju Jaya" />,
};
export const ContextPanelLongTitle: Story = {
	render: (args) => <ContextShell {...args} title="Organisasi Dagang Sejahtera Makmur Abadi Nusantara Raya Indonesi" />,
};

/** Linear-style: content in a rounded panel on the sidebar surface; toggle and search live in the sidebar. */
export const Inset: Story = { ...NavigationAndSearch, args: { inset: true } };
export const InsetCollapsed: Story = { ...NavigationAndSearch, args: { inset: true, defaultCollapsed: true } };
/** The panel, not the window, scrolls on desktop; the slim header stays put. */
export const InsetLongContent: Story = { ...LongContent, args: { ...LongContent.args, inset: true } };
export const InsetCustomHeader: Story = { ...WorkspaceAndAccount, args: { inset: true } };
export const InsetMobile: Story = { ...Inset, globals: { viewport: { value: "mobile1", isRotated: false } } };
