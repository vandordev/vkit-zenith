import type { Meta, StoryObj } from "@storybook/react";
import { CreditCard, LayoutDashboard, Package, Settings } from "lucide-react";
import { useState } from "react";

import { Button } from "./button";
import { CommandPalette, type CommandPaletteItem } from "./command-palette";

const items: CommandPaletteItem[] = [
	{ id: "overview", label: "Overview", group: "Pages", icon: <LayoutDashboard /> },
	{ id: "payments", label: "Payments", group: "Pages", icon: <CreditCard /> },
	{ id: "products", label: "Products", group: "Pages", icon: <Package /> },
	{ id: "settings", label: "Workspace settings", group: "Settings", icon: <Settings />, hint: "Settings" },
	...Array.from({ length: 12 }, (_, index) => ({ id: `org-${index}`, label: `Organization ${index + 1}`, group: "Switch organization" })),
];

const Preview = ({ defaultOpen = true, list = items }: { defaultOpen?: boolean; list?: CommandPaletteItem[] }) => {
	const [open, setOpen] = useState(defaultOpen);
	const [selected, setSelected] = useState<string>();
	return (
		<div className="flex flex-col items-start gap-3 p-6">
			<Button variant="outline" onClick={() => setOpen(true)}>
				Open palette
			</Button>
			<p className="text-sm text-muted-foreground" aria-live="polite">
				{selected ? `Selected: ${selected}` : "Nothing selected"}
			</p>
			<CommandPalette open={open} onOpenChange={setOpen} items={list} onSelect={(item) => setSelected(item.label)} />
		</div>
	);
};

const meta = {
	title: "UI/CommandPalette",
	component: CommandPalette,
	parameters: { layout: "fullscreen" },
	render: () => <Preview />,
	args: { open: true, onOpenChange: () => {}, items, onSelect: () => {} },
	argTypes: { items: { control: false }, onOpenChange: { control: false }, onSelect: { control: false } },
} satisfies Meta<typeof CommandPalette>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Closed: Story = { render: () => <Preview defaultOpen={false} /> };
export const Empty: Story = { render: () => <Preview list={[]} /> };
