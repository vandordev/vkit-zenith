import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { RadioGroup, RadioGroupItem } from "./radio-group";

const options = [
	{ label: "Transfer bank", value: "bank" },
	{ label: "QRIS", value: "qris" },
	{ label: "Kartu kredit", value: "card" },
];

const meta = {
	argTypes: {
		disabled: { control: "boolean" },
		onValueChange: { control: false },
		readOnly: { control: "boolean" },
		render: { control: false },
		value: { control: false },
	},
	args: { "aria-label": "Metode pembayaran", defaultValue: "qris", disabled: false, readOnly: false },
	component: RadioGroup,
	parameters: { layout: "centered" },
	render: ({ ...args }) => (
		<RadioGroup {...args}>
			{options.map((option) => (
				<label key={option.value} className="flex items-center gap-3 text-sm">
					<RadioGroupItem value={option.value} />
					<span>{option.label}</span>
				</label>
			))}
		</RadioGroup>
	),
	title: "UI/Radio Group",
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;
export const Playground: Story = {};
export const Disabled: Story = { args: { disabled: true } };
export const ReadOnly: Story = { args: { readOnly: true } };
export const Invalid: Story = {
	render: (args) => (
		<RadioGroup {...args} aria-invalid>
			{options.map((option) => (
				<label key={option.value} className="flex items-center gap-3 text-sm">
					<RadioGroupItem aria-invalid value={option.value} />
					<span>{option.label}</span>
				</label>
			))}
		</RadioGroup>
	),
};
export const WithoutMotion: Story = {
	render: (args) => (
		<RadioGroup {...args}>
			{options.map((option) => (
				<label key={option.value} className="flex items-center gap-3 text-sm">
					<RadioGroupItem animated={false} value={option.value} />
					<span>{option.label}</span>
				</label>
			))}
		</RadioGroup>
	),
};

const ControlledExample = () => {
	const [value, setValue] = useState("bank");
	return (
		<div className="grid gap-3">
			<RadioGroup aria-label="Metode pembayaran" value={value} onValueChange={setValue}>
				{options.map((option) => (
					<label key={option.value} className="flex items-center gap-3 text-sm">
						<RadioGroupItem value={option.value} />
						<span>{option.label}</span>
					</label>
				))}
			</RadioGroup>
			<p className="text-sm text-muted-foreground">Dipilih: {value}</p>
		</div>
	);
};
export const Controlled: Story = { render: () => <ControlledExample /> };
