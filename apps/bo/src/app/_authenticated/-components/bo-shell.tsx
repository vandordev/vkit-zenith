import type { ReactNode } from "react";
import { brand } from "@repo/brand";
import { ZenithMark, ZenithWordmark } from "@repo/ui/zenith-mark";
import { AppShell, type AppShellNavigationItem, type AppShellProps } from "@repo/ui/app-shell";
import type { ThemePreference } from "../../../lib/theme";
import { NavUser } from "./nav-user";
import type { BoV1GetMeResponse } from "../../../lib/api/generated/types.gen";

export function BOShell({
	staff,
	navigation,
	navigationContext,
	sidebarContext,
	children,
	onLogout,
	onAccount,
	theme,
	onThemeChange,
	pending,
}: {
	staff: BoV1GetMeResponse;
	navigation: AppShellNavigationItem[];
	/** Panel identity for the drill-down slide; omit for a single static panel. */
	navigationContext?: AppShellProps["navigationContext"];
	sidebarContext?: AppShellProps["sidebarContext"];
	children: ReactNode;
	onLogout: () => void;
	onAccount: () => void;
	theme: ThemePreference;
	onThemeChange: (theme: ThemePreference) => void;
	pending?: boolean;
}) {
	return (
		<AppShell
			showHeader={false}
			inset
			labels={{
				skipToContent: "Lewati ke konten",
				sidebar: "Navigasi samping",
				openNavigation: "Buka navigasi",
				closeNavigation: "Tutup navigasi",
				expandSidebar: "Perluas navigasi",
				collapseSidebar: "Ciutkan navigasi",
			}}
			brand={{
				name: `${brand.name} Backoffice`,
				icon: <ZenithMark size={28} />,
				label: <ZenithWordmark surface="Backoffice" />,
			}}
			navigation={navigation}
			navigationContext={navigationContext}
			sidebarContext={sidebarContext}
			navigationLabel="Navigasi backoffice"
			sidebarFooter={({ collapsed, closeNavigation }) => (
				<NavUser
					staff={staff}
					collapsed={collapsed}
					pending={pending}
					theme={theme}
					onThemeChange={onThemeChange}
					onAccount={() => {
						closeNavigation();
						onAccount();
					}}
					onLogout={onLogout}
				/>
			)}
		>
			{children}
		</AppShell>
	);
}
