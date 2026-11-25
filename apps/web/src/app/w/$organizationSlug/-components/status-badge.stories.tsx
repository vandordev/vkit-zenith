import type { Meta, StoryObj } from "@storybook/react";
import { StatusBadge } from "./status-badge";

const meta = {
	title: "Web/Components/Dashboard/StatusBadge",
	component: StatusBadge,
	args: { resource: "apiKey", status: "active" },
} satisfies Meta<typeof StatusBadge>;
export default meta;
type Story = StoryObj<typeof meta>;

export const EveryStatus: Story = {
	render: () => (
		<div className="flex flex-col gap-3">
			{(
				[
					["organization", ["active", "suspended"]],
					["apiKey", ["active", "expired", "revoked"]],
				] as const
			).map(([resource, statuses]) => (
				<div key={resource} className="flex flex-wrap items-center gap-2">
					<span className="w-28 font-mono text-xs text-muted-foreground">{resource}</span>
					{statuses.map((status) => (
						<StatusBadge key={status} resource={resource} status={status} />
					))}
				</div>
			))}
		</div>
	),
};
/** A status from a newer API renders neutral with its raw value. */
export const UnknownStatus: Story = { args: { resource: "apiKey", status: "rotating" } };
