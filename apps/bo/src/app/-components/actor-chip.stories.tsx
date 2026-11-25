import type { Meta, StoryObj } from "@storybook/react";
import { ActorChip } from "./actor-chip";
import { actors } from "../-fixtures/audit";

const meta = {
	title: "BO/Components/ActorChip",
	component: ActorChip,
	args: { actor: actors.staff },
	argTypes: { actor: { control: false } },
} satisfies Meta<typeof ActorChip>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Staff: Story = {};
export const User: Story = { args: { actor: actors.user } };
export const Client: Story = { args: { actor: actors.client } };
export const System: Story = { args: { actor: actors.system } };
export const Anonymous: Story = { args: { actor: actors.anonymous } };
export const UnknownSystem: Story = { args: { actor: actors.unknownSystem } };
export const UnknownKind: Story = {
	args: { actor: { ...actors.system, kind: "robot" as never, system: null } },
	parameters: { docs: { description: { story: "A kind from a newer API renders its raw value instead of crashing." } } },
};
export const LongName: Story = {
	args: { actor: actors.longName },
	decorators: [
		(Story) => (
			<div className="max-w-xs">
				<Story />
			</div>
		),
	],
};
