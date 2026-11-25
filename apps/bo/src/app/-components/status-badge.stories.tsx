import type { Meta, StoryObj } from "@storybook/react";
import { StatusBadge, type StatusKind } from "./status-badge";

const samples: Record<StatusKind, string[]> = {
	apiKey: ["active", "expired", "revoked"],
	job: ["available", "scheduled", "running", "retryable", "completed", "cancelled", "discarded"],
	session: ["active", "expired", "revoked"],
	emailAction: ["pending", "consumed", "revoked", "expired"],
	audit: ["success", "failure", "denied"],
};
const meta = {
	title: "BO/Components/StatusBadge",
	component: StatusBadge,
	args: { kind: "apiKey", status: "active" },
} satisfies Meta<typeof StatusBadge>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const EveryStatus: Story = {
	render: () => (
		<dl className="grid gap-3 sm:grid-cols-[10rem_1fr]">
			{(Object.keys(samples) as StatusKind[]).map((kind) => (
				<div key={kind} className="contents">
					<dt className="font-mono text-xs text-muted-foreground">{kind}</dt>
					<dd className="flex flex-wrap gap-2">
						{samples[kind].map((status) => (
							<StatusBadge key={status} kind={kind} status={status} />
						))}
					</dd>
				</div>
			))}
		</dl>
	),
};
export const UnknownStatus: Story = { args: { kind: "job", status: "paused_by_operator" } };
export const LongUnknownStatus: Story = {
	args: { kind: "session", status: "status_baru_dari_versi_api_yang_lebih_baru_dengan_nama_sangat_panjang" },
	decorators: [
		(Story) => (
			<div className="max-w-48">
				<Story />
			</div>
		),
	],
};
