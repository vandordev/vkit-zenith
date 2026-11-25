import type { Meta, StoryObj } from "@storybook/react";
import { useRef, useState } from "react";

import { Button } from "./button";
import { Input } from "./input";
import type { InputProps } from "./input";

const meta = {
	argTypes: {
		"aria-invalid": { control: "boolean" },
		defaultValue: { control: "text" },
		disabled: { control: "boolean" },
		icon: { control: false },
		label: { control: "text" },
		labelStyle: { control: "select", options: ["floating", "static"] },
		onChange: { control: false },
		placeholder: { control: "text" },
		readOnly: { control: "boolean" },
		type: { control: "select", options: ["text", "email", "url", "tel"] },
		value: { control: false },
	},
	args: {
		"aria-invalid": false,
		defaultValue: "",
		disabled: false,
		label: "Full name",
		labelStyle: "floating",
		placeholder: "Alex Morgan",
		readOnly: false,
		type: "text",
	},
	component: Input,
	decorators: [
		(Story) => (
			<div className="w-full max-w-sm">
				<Story />
			</div>
		),
	],
	parameters: { layout: "padded" },
	render: (args) => <Input key={String(args.defaultValue)} {...args} />,
	title: "UI/Input",
} satisfies Meta<typeof Input>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Playground: Story = {};
export const StaticLabel: Story = { args: { labelStyle: "static" } };
export const WithIcon: Story = {
	args: {
		icon: (
			<svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
				<rect x="3" y="5" width="18" height="14" rx="2" />
				<path d="m3 6 9 7 9-7" />
			</svg>
		),
		label: "Email",
		placeholder: "alex@example.com",
		type: "email",
	},
};
export const Populated: Story = { args: { defaultValue: "Alex Morgan" } };
export const Disabled: Story = {
	args: { defaultValue: "Alex Morgan", disabled: true },
};
export const ReadOnly: Story = {
	args: { defaultValue: "Alex Morgan", readOnly: true },
};
export const Invalid: Story = {
	args: { "aria-describedby": "name-error", "aria-invalid": true },
	render: (args) => (
		<div className="grid gap-2">
			<Input {...args} />
			<p id="name-error" className="text-sm text-destructive">
				Enter your full name.
			</p>
		</div>
	),
};
const ControlledField = ({ defaultValue, onChange, value: _value, ...args }: InputProps) => {
	const [value, setValue] = useState(String(defaultValue ?? ""));
	return (
		<div className="grid gap-2">
			<Input
				{...args}
				value={value}
				onChange={(event) => {
					setValue(event.currentTarget.value);
					onChange?.(event);
				}}
			/>
			<output className="text-xs text-muted-foreground">Value: {value || "empty"}</output>
		</div>
	);
};
export const Controlled: Story = {
	argTypes: { defaultValue: { control: false } },
	render: (args) => <ControlledField {...args} />,
};

export const File: Story = {
	args: {
		label: "Choose a .txt or .md document",
		labelStyle: "static",
		type: "file",
		accept: ".txt,.md,text/plain,text/markdown",
		defaultValue: undefined,
		placeholder: undefined,
	},
	render: function FilePreview(args) {
		const [filename, setFilename] = useState("No file selected");
		return (
			<div className="grid gap-3">
				<Input {...args} onChange={(event) => setFilename(event.currentTarget.files?.[0]?.name ?? "No file selected")} />
				<p role="status">{filename}</p>
			</div>
		);
	},
};

export const Form: Story = {
	render: function FormPreview() {
		const [saved, setSaved] = useState("");
		const ref = useRef<HTMLInputElement>(null);
		return (
			<form
				className="grid max-w-md gap-4"
				onSubmit={(event) => {
					event.preventDefault();
					setSaved(String(new FormData(event.currentTarget).get("title")));
				}}
			>
				<Input
					ref={ref}
					id="fixture-title"
					label="Document title"
					labelStyle="static"
					name="title"
					required
					defaultValue="Quarterly report"
				/>
				<Button type="submit">Save title</Button>
				<Button type="button" variant="outline" onClick={() => ref.current?.focus()}>
					Focus title
				</Button>
				<p role="status">{saved ? `Saved: ${saved}` : "Local form fixture"}</p>
			</form>
		);
	},
};
