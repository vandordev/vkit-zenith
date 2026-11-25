import type { ReactNode } from "react";
import { Link, type LinkProps } from "@tanstack/react-router";
import {
	ArrowLeft,
	Building2,
	Contact,
	History,
	ListChecks,
	House,
	KeyRound,
	LayoutDashboard,
	MonitorSmartphone,
	UsersRound,
	type LucideIcon,
} from "lucide-react";
import type { AppShellNavigationItem, AppShellProps } from "@repo/ui/app-shell";
import { buttonVariants } from "@repo/ui/button";
import { type NavIcon, type NavItem, type ResolvedNavigation } from "../-navigation";

const icons: Record<NavIcon, LucideIcon> = {
	home: House,
	organizations: Building2,
	users: Contact,
	staff: UsersRound,
	activity: History,
	sessions: MonitorSmartphone,
	overview: LayoutDashboard,
	members: UsersRound,
	"api-keys": KeyRound,
	jobs: ListChecks,
};
// `to` is a typed route pattern (FileRouteTypes); the union is too wide for Link's per-route inference.
const linkProps = (item: Pick<NavItem, "to" | "params" | "search">) =>
	({ to: item.to, params: item.params, search: item.search }) as LinkProps;
// Labels stay laid out and fade, so collapsing the rail never re-centres the icons.
const fadeClass = "transition-opacity duration-(--motion-fast) ease-(--ease-out) motion-reduce:transition-none";
const itemClass =
	"flex h-8 w-full items-center overflow-hidden rounded-md py-1.5 text-xs transition-colors hover:bg-accent max-md:h-10 aria-[current=page]:bg-accent aria-[current=page]:font-semibold";

export function buildShellNavigation(items: readonly NavItem[]): AppShellNavigationItem[] {
	return items.map((item) => {
		const Icon = icons[item.icon];
		return {
			id: item.id,
			label: item.label,
			active: item.active,
			dividerAfter: item.dividerAfter,
			renderControl: ({ closeNavigation, collapsed }) => (
				<Link
					{...linkProps(item)}
					// Active state comes from resolveNavigation; Link's prefix matching would also mark ancestors.
					activeOptions={{ exact: true }}
					onClick={closeNavigation}
					aria-label={item.label}
					aria-current={item.active ? "page" : undefined}
					className={`${itemClass} gap-2 px-2`}
				>
					<Icon aria-hidden="true" className="size-4 shrink-0" />
					<span className={`${fadeClass} min-w-0 truncate${collapsed ? " opacity-0" : ""}`}>{item.label}</span>
				</Link>
			),
		};
	});
}

export function buildSidebarContext(data: ResolvedNavigation["sidebarContext"]): AppShellProps["sidebarContext"] {
	if (!data) return undefined;
	return {
		title: data.title,
		back: {
			label: data.back.label,
			render: ({ collapsed, closeNavigation }): ReactNode => (
				// A real link (not role=button) styled as the shared ghost button.
				<Link
					to={data.back.to}
					params={data.back.params}
					search={data.back.search as LinkProps["search"]}
					onClick={closeNavigation}
					aria-label={data.back.label}
					// Full row width like the navigation items, so the whole row is clickable.
					className={`${buttonVariants({ variant: "ghost", size: "sm" })} w-full justify-start has-[>svg]:px-2!`}
				>
					<ArrowLeft aria-hidden="true" />
					<span className={`${fadeClass} truncate${collapsed ? " opacity-0" : ""}`}>{data.back.label}</span>
				</Link>
			),
		},
	};
}
