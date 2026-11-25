import type { Meta, StoryObj } from "@storybook/react";
import { cn } from "cn";
import { Plus } from "lucide-react";

import { Button } from "./button";
import { Checkbox } from "./checkbox";
import { Input } from "./input";
import { Select } from "./select";
import { Switch } from "./switch";
import { TextArea } from "./textarea";

const currencies = [
	{ label: "IDR — Rupiah", value: "idr" },
	{ label: "USD — US Dollar", value: "usd" },
];

const Panel = ({ dark }: { dark?: boolean }) => (
	<div className={cn("grid gap-6 rounded-xl bg-background p-6 text-foreground", dark && "dark")}>
		<p className="text-sm font-medium text-muted-foreground">{dark ? "Dark" : "Light"}</p>
		<div className="flex flex-wrap items-center gap-3">
			<Button>
				<Plus /> Buat payment link
			</Button>
			<Button variant="secondary">Secondary</Button>
			<Button variant="outline">Outline</Button>
			<Button variant="ghost">Ghost</Button>
			<Button variant="destructive">Hapus</Button>
		</div>
		<div className="grid gap-4 rounded-lg border bg-card p-4">
			<Input label="Nama produk" labelStyle="static" placeholder="Langganan Pro" />
			<Input label="Email pelanggan" defaultValue="budi@contoh.id" />
			<Select data={currencies} placeholder="Pilih mata uang" aria-label="Mata uang" className="w-full" />
			<TextArea label="Catatan" placeholder="Opsional" />
			<div className="flex items-center gap-6 text-sm">
				<span className="flex items-center gap-2">
					<Checkbox aria-label="Kirim struk" defaultChecked /> Kirim struk
				</span>
				<span className="flex items-center gap-2">
					<Checkbox aria-label="Simpan draf" /> Simpan draf
				</span>
				<span className="flex items-center gap-2">
					<Switch aria-label="Aktif" /> Aktif
				</span>
			</div>
			<div className="flex justify-end gap-2">
				<Button variant="outline">Batal</Button>
				<Button>Simpan</Button>
			</div>
		</div>
	</div>
);

const meta = {
	parameters: { layout: "fullscreen" },
	title: "UI/Control Surfaces",
} satisfies Meta;

export default meta;

/** Buttons and fields together in both themes, to judge clickability against input surfaces. */
export const LightAndDark: StoryObj = {
	render: () => (
		<div className="grid gap-4 p-4 md:grid-cols-2">
			<Panel />
			<Panel dark />
		</div>
	),
};
