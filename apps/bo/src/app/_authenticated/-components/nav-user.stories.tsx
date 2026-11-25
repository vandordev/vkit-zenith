import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { NavUser } from "./nav-user";
import { staff, noop } from "../../-fixtures";

const meta = {
	title: "BO/Layouts/NavUser",
	component: NavUser,
	parameters: { layout: "centered" },
	args: { staff, theme: "system", onThemeChange: noop, onAccount: noop, onLogout: noop },
	argTypes: { staff: { control: false }, theme: { control: "inline-radio", options: ["light", "dark", "system"] }, onThemeChange: { control: false }, onAccount: { control: false }, onLogout: { control: false } },
	render: function Preview(args) {
		const [action, setAction] = useState("");
		const [theme, setTheme] = useState(args.theme);
		return (
			<div className="flex flex-col gap-3">
				<div className={args.collapsed ? "w-12" : "w-56"}>
					<NavUser
						{...args}
						theme={theme}
						onThemeChange={setTheme}
						onAccount={() => setAction("Pratinjau Akun Saya, tanpa navigasi.")}
						onLogout={() => setAction("Pratinjau keluar, tanpa panggilan API.")}
					/>
				</div>
				<p role="status" className="text-xs text-muted-foreground">
					{action}
				</p>
			</div>
		);
	},
} satisfies Meta<typeof NavUser>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Collapsed: Story = { args: { collapsed: true } };
export const Pending: Story = { args: { pending: true } };
export const LongName: Story = { args: { staff: { ...staff, displayName: "Nadia Prameswari dengan nama panjang untuk tim operasional" } } };
