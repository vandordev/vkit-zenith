import type { Meta, StoryObj } from "@storybook/react";
import { useRef, useState } from "react";
import { Button } from "@repo/ui/button";
import { ActivityDetail } from "./activity-detail";
import { detail, summary, missing, noop, customerDetail } from "../-fixtures";
import { activityTarget } from "@bo/lib/api/activity-navigation";
const meta = {
	title: "BO/Pages/Activity/Detail",
	component: ActivityDetail,
	args: { selected: detail.id, data: detail, onClose: noop, onRetry: noop },
	argTypes: { onClose: { control: false }, onRetry: { control: false }, targetLink: { control: false }, returnFocus: { control: false } },
	parameters: { layout: "padded" },
} satisfies Meta<typeof ActivityDetail>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {
	render: function Preview(args) {
		const [selected, setSelected] = useState<string>();
		const trigger = useRef<HTMLButtonElement>(null);
		return (
			<>
				<Button ref={trigger} onClick={() => setSelected(args.selected)}>
					Periksa aktivitas
				</Button>
				<ActivityDetail {...args} selected={selected} onClose={() => setSelected(undefined)} returnFocus={trigger} />
			</>
		);
	},
};
export const DirectEntry: Story = {
	render: function Preview(args) {
		const [selected, setSelected] = useState(args.selected);
		const fallback = useRef<HTMLDivElement>(null);
		return (
			<>
				<div ref={fallback} tabIndex={-1}>
					Daftar aktivitas
				</div>
				<ActivityDetail {...args} selected={selected} onClose={() => setSelected(undefined)} returnFocus={fallback} />
			</>
		);
	},
};
export const AnonymousSummary: Story = {
	...DirectEntry,
	args: { selected: summary.id, data: { ...summary, detailAvailable: false, changedFields: [], previous: null, current: null } },
};
export const MissingTarget: Story = {
	...DirectEntry,
	args: { selected: missing.id, data: { ...missing, detailAvailable: false, changedFields: [], previous: null, current: null } },
};
export const RecordedStaffState: Story = {
	...DirectEntry,
	args: {
		data: {
			...detail,
			action: "staff.update",
			target: {
				type: "staff",
				id: "00000000-0000-4000-8000-000000000021",
				name: "Raka Putra",
				organizationId: null,
				available: true,
			},
			previous: { role: "admin", status: "active" },
			current: { role: "superadmin", status: "disabled" },
			changedFields: ["role", "status"],
		},
	},
};
export const UnavailableDetail: Story = { ...DirectEntry, args: { data: { ...detail, detailAvailable: false, changedFields: [] } } };
export const Pending: Story = { ...DirectEntry, args: { data: undefined, loading: true } };
export const Error: Story = { ...DirectEntry, args: { data: undefined, error: "Aktivitas tidak tersedia." } };
export const SwitchingRecord: Story = { ...DirectEntry, args: { selected: "00000000-0000-4000-8000-000000000099" } };
export const CustomerMembership: Story = {
	...DirectEntry,
	args: {
		selected: customerDetail.id,
		data: customerDetail,
		targetLink: (
			<a className="underline" href={activityTarget(customerDetail)}>
				Buka target
			</a>
		),
	},
};
export const CustomerEmail: Story = {
	...DirectEntry,
	args: {
		selected: customerDetail.id,
		data: {
			...customerDetail,
			action: "customer.user.email_change",
			target: {
				...customerDetail.target,
				type: "user",
				organizationId: null,
				id: "00000000-0000-4000-8000-000000000201",
				name: "Nadia Prameswari",
			},
			changedFields: ["email", "email_verified_at"],
		},
		targetLink: (
			<a className="underline" href="/users/00000000-0000-4000-8000-000000000201">
				Buka target
			</a>
		),
	},
};
export const CustomerUnavailable: Story = {
	...DirectEntry,
	args: {
		selected: customerDetail.id,
		data: { ...customerDetail, target: { ...customerDetail.target, available: false, organizationId: null, name: null } },
		targetLink: undefined,
	},
};
