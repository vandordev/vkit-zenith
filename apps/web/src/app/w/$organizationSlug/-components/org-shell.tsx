import { useState, type ReactNode } from "react";
import { Link } from "@tanstack/react-router";
import { ArrowLeft, Building2, KeyRound, LayoutDashboard, Settings, UsersRound, type LucideIcon } from "lucide-react";
import { brand } from "@repo/brand";
import { AppShell, type AppShellNavigationItem, type AppShellProps } from "@repo/ui/app-shell";
import { buttonVariants } from "@repo/ui/button";
import { CommandPalette } from "@repo/ui/command-palette";
import { ZenithMark, ZenithWordmark } from "@repo/ui/zenith-mark";
import type { V1MyOrganization } from "../../../../lib/api/generated/types.gen";
import type { ThemePreference } from "../../../../lib/theme";
import { OrganizationBanners } from "./org-banners";
import { NavUser } from "./nav-user";
import { OrgSwitcher } from "./org-switcher";
import { groupLabels, linkProps, resolveCommandItems, type OrgNavigation, type RouteLink } from "./navigation";

const icons: Record<string, LucideIcon> = {
	overview: LayoutDashboard,
	settings: Settings,
	profile: Building2,
	members: UsersRound,
	"api-keys": KeyRound,
};
const roleLabels: Record<V1MyOrganization["role"], string> = { owner: "Owner", admin: "Admin", member: "Anggota" };
// Labels stay laid out and fade, so collapsing the rail never re-centres the icons.
const fadeClass = "transition-opacity duration-(--motion-fast) ease-(--ease-out) motion-reduce:transition-none";
const itemClass =
	"flex h-8 w-full items-center overflow-hidden rounded-md py-1.5 text-xs transition-colors hover:bg-accent max-md:h-10 aria-[current=page]:bg-accent aria-[current=page]:font-semibold";
function buildItems(navigation: OrgNavigation): AppShellNavigationItem[] {
	const result: AppShellNavigationItem[] = [];
	let group: string | undefined;
	navigation.items.forEach((item, index) => {
		if (item.group && item.group !== group) {
			const label = groupLabels[item.group];
			// A heading, not a button; the collapsed rail shows a divider instead.
			result.push({
				id: `group-${item.group}`,
				label,
				tooltip: false,
				renderControl: ({ collapsed }) =>
					collapsed ? (
						<div role="separator" className="my-2 border-b border-border" />
					) : (
						<p className="px-2 pt-3 pb-1 text-xs font-medium text-muted-foreground">{label}</p>
					),
			});
		}
		group = item.group;
		const Icon = icons[item.id] ?? LayoutDashboard;
		result.push({
			id: item.id,
			label: item.label,
			active: item.active,
			// Pengaturan is last and set apart.
			dividerAfter: navigation.items[index + 1]?.id === "settings",
			renderControl: ({ closeNavigation, collapsed }) => (
				<Link
					{...linkProps(item)}
					activeOptions={{ exact: true }}
					onClick={closeNavigation}
					aria-label={item.label}
					aria-current={item.active ? "page" : undefined}
					className={`${itemClass} gap-2 px-2`}
				>
					<Icon aria-hidden="true" className="size-4 shrink-0" />
					<span className={`${fadeClass} min-w-0 flex-1 truncate${collapsed ? " opacity-0" : ""}`}>{item.label}</span>
				</Link>
			),
		});
	});
	return result;
}

function settingsContext(navigation: OrgNavigation): AppShellProps["sidebarContext"] {
	if (navigation.panel !== "settings") return undefined;
	const first = navigation.items[0];
	if (!first) return undefined;
	return {
		title: "Pengaturan",
		back: {
			label: "Kembali",
			render: ({ collapsed, closeNavigation }) => (
				<Link
					{...linkProps({ to: "/w/$organizationSlug", params: first.params })}
					onClick={closeNavigation}
					aria-label="Kembali"
					className={`${buttonVariants({ variant: "ghost", size: "sm" })} w-full justify-start has-[>svg]:px-2!`}
				>
					<ArrowLeft aria-hidden="true" />
					<span className={`${fadeClass} truncate${collapsed ? " opacity-0" : ""}`}>Kembali</span>
				</Link>
			),
		},
	};
}

export function OrgShell({
	organization,
	organizations,
	defaultCollapsed,
	navigation,
	userName,
	logoutPending,
	onLogout,
	theme,
	onThemeChange,
	onNavigate,
	children,
}: {
	organization: V1MyOrganization;
	organizations: readonly V1MyOrganization[];
	defaultCollapsed?: boolean;
	navigation: OrgNavigation;
	userName: string;
	logoutPending?: boolean;
	onLogout: () => void;
	theme: ThemePreference;
	onThemeChange: (theme: ThemePreference) => void;
	/** Command palette selection; the route adapter owns navigation. */
	onNavigate: (link: RouteLink) => void;
	children: ReactNode;
}) {
	const current = navigation.items.find((item) => item.active);
	const [searchOpen, setSearchOpen] = useState(false);
	const commands = resolveCommandItems({ slug: organization.slug, organizations }).map((item) => {
		const Icon = icons[item.id] ?? Building2;
		return { ...item, icon: <Icon /> };
	});
	return (
		<AppShell
			inset
			onSearch={() => setSearchOpen(true)}
			searchLabel="Cari halaman"
			labels={{
				skipToContent: "Lewati ke konten",
				sidebar: "Navigasi samping",
				openNavigation: "Buka navigasi",
				closeNavigation: "Tutup navigasi",
				expandSidebar: "Perluas navigasi",
				collapseSidebar: "Ciutkan navigasi",
			}}
			brand={{ name: brand.name, icon: <ZenithMark size={28} />, label: <ZenithWordmark /> }}
			navigation={buildItems(navigation)}
			navigationContext={{ key: navigation.panel, depth: navigation.panel === "settings" ? 1 : 0 }}
			sidebarContext={settingsContext(navigation)}
			defaultCollapsed={defaultCollapsed}
			navigationLabel="Navigasi dashboard"
			sidebarHeader={({ collapsed }) => <OrgSwitcher organization={organization} organizations={organizations} collapsed={collapsed} />}
			// The organization is already named by the switcher right beside it.
			header={current && <p className="min-w-0 truncate text-sm font-medium">{current.label}</p>}
			sidebarFooter={({ collapsed }) => (
				<NavUser
					name={userName}
					role={roleLabels[organization.role]}
					collapsed={collapsed}
					pending={logoutPending}
					theme={theme}
					onThemeChange={onThemeChange}
					onLogout={onLogout}
				/>
			)}
		>
			<div className="flex flex-col gap-4">
				<div className="px-4 pt-4 sm:px-6 empty:hidden">
					<OrganizationBanners suspended={organization.status === "suspended"} />
				</div>
				{children}
			</div>
			<CommandPalette
				open={searchOpen}
				onOpenChange={setSearchOpen}
				items={commands}
				onSelect={(item) => onNavigate(item.link)}
				labels={{
					title: "Cari halaman",
					placeholder: "Cari halaman atau organisasi…",
					empty: "Tidak ada yang cocok.",
					close: "Tutup pencarian",
					activate: "Buka",
					instructions: "Gunakan panah atas dan bawah untuk memilih, lalu Enter untuk membuka.",
				}}
			/>
		</AppShell>
	);
}
