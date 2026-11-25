import { ChevronsUpDown, CircleUserRound, LogOut, Monitor, Moon, Sun } from "lucide-react";
import { Blobatar } from "@repo/ui/blobatar";
import { Button } from "@repo/ui/button";
import { Dropdown } from "@repo/ui/dropdown";
import { themePreferences, type ThemePreference } from "../../../lib/theme";
import type { BoV1GetMeResponse } from "../../../lib/api/generated/types.gen";

const themeIcons = { light: Sun, dark: Moon, system: Monitor } satisfies Record<ThemePreference, unknown>;

export function NavUser({
	staff,
	collapsed = false,
	pending,
	theme,
	onThemeChange,
	onAccount,
	onLogout,
}: {
	staff: BoV1GetMeResponse;
	collapsed?: boolean;
	pending?: boolean;
	theme: ThemePreference;
	onThemeChange: (theme: ThemePreference) => void;
	onAccount: () => void;
	onLogout: () => void;
}) {
	const role = staff.role === "superadmin" ? "Superadmin" : "Admin";
	return (
		<Dropdown
			disabled={pending}
			contentProps={{ side: collapsed ? "right" : "top", align: "end", className: "w-56" }}
			trigger={
				<Button
					variant="ghost"
					aria-label={`Menu pengguna: ${staff.displayName}`}
					aria-busy={pending || undefined}
					// The 28px avatar sits on the sidebar's 24px icon axis (the 48px rail's centre) in both states, so
					// collapsing never moves it; only the text fades.
					className="flex h-auto min-h-10 w-full justify-start gap-2 overflow-hidden ps-0.5 pe-2 py-2 has-[>svg]:ps-0.5 has-[>svg]:pe-2"
				>
					<Blobatar name={staff.id} alt="" size={28} className="shrink-0" />
					<span
						aria-hidden={collapsed || undefined}
						className={`flex min-w-0 flex-1 items-center gap-2 transition-opacity duration-(--motion-fast) motion-reduce:transition-none${collapsed ? " opacity-0" : ""}`}
					>
						<span className="min-w-0 flex-1 text-left text-xs">
							<span className="block truncate font-medium">{staff.displayName}</span>
							<span className="block truncate font-normal text-muted-foreground">{role}</span>
						</span>
						<ChevronsUpDown aria-hidden="true" className="size-3.5 shrink-0" />
					</span>
				</Button>
			}
			items={[
				{
					id: "user",
					type: "group",
					label: `${staff.displayName} · ${role}`,
					items: [
						{ id: "account", label: "Akun Saya", icon: CircleUserRound, onSelect: onAccount },
						{
							id: "theme",
							type: "submenu",
							label: "Tema",
							icon: themeIcons[theme],
							items: [
								{
									id: "theme-choice",
									type: "radio-group",
									value: theme,
									onValueChange: (value) => {
										const next = themePreferences.find((preference) => preference === value);
										if (next) onThemeChange(next);
									},
									items: [
										{ id: "theme-light", value: "light", label: "Terang", icon: Sun },
										{ id: "theme-dark", value: "dark", label: "Gelap", icon: Moon },
										{ id: "theme-system", value: "system", label: "Sistem", icon: Monitor },
									],
								},
							],
						},
					],
				},
				{ id: "separator", type: "separator" },
				{ id: "logout", label: "Keluar", icon: LogOut, onSelect: onLogout },
			]}
		/>
	);
}
