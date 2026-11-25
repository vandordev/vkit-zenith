import type { Meta, StoryObj } from "@storybook/react";
import { ZenithLoader } from "./zenith-loader";
import { ZenithMark } from "./zenith-mark";

const meta = {
	title: "UI/ZenithLoader",
	component: ZenithLoader,
	parameters: { layout: "fullscreen" },
	args: { mode: "screen" },
} satisfies Meta<typeof ZenithLoader>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Screen: Story = {};
export const DrawAndFill: Story = { args: { size: 160, hideLabel: false } };
export const WithLabel: Story = { args: { hideLabel: false, label: "Memuat…" } };
export const Container: Story = {
	args: { mode: "container", label: "Memuat data…", size: 56 },
	decorators: [
		(Story) => (
			<div className="mx-auto h-80 max-w-lg">
				<Story />
			</div>
		),
	],
};
export const IconOnly: Story = { ...Container, args: { mode: "container", hideLabel: true, size: 40 } };
export const Multiple: Story = {
	args: { mode: "container" },
	render: () => (
		<div className="grid grid-cols-2 gap-4">
			<ZenithLoader label="Memuat data…" />
			<ZenithLoader label="Memuat data…" size={48} />
		</div>
	),
};
export const Dark: Story = {
	args: { mode: "container", hideLabel: false, size: 96 },
	decorators: [
		(Story) => (
			<div className="dark h-80 bg-background text-foreground">
				<Story />
			</div>
		),
	],
};
export const MarkLightAndDark: Story = {
	render: () => (
		<div className="flex gap-6 p-6">
			<ZenithMark size={64} title="Zenith" />
			<div className="dark rounded-lg bg-background p-3">
				<ZenithMark size={64} title="Zenith" />
			</div>
		</div>
	),
};
