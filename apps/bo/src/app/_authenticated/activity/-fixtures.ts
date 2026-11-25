import type { V1ActivityRow, V1ActivityDetail, V1ActivityAction, V1ActivityListBody } from "@bo/lib/api/generated/types.gen";
export const noop = () => {};
export const actions: V1ActivityAction[] = [
	{ value: "staff.login", label: "Staf masuk" },
	{ value: "staff.update", label: "Staf diperbarui" },
	{ value: "customer.user.email_change", label: "Email pelanggan diganti" },
	{ value: "customer.membership.role", label: "Kepemilikan organisasi diubah" },
];
export const event: V1ActivityRow = {
	id: "00000000-0000-4000-8000-000000000011",
	occurredAt: "2026-10-26T09:10:00Z",
	action: "staff.update",
	outcome: "success",
	actor: { kind: "staff", id: "00000000-0000-4000-8000-000000000001", name: "Nadia Prameswari" },
	target: {
		type: "staff",
		id: "00000000-0000-4000-8000-000000000003",
		name: "Raka Putra",
		organizationId: null,
		available: true,
	},
};
export const detail: V1ActivityDetail = {
	...event,
	detailAvailable: true,
	changedFields: ["role", "status"],
	previous: null,
	current: null,
};
export const summary: V1ActivityRow = {
	...event,
	id: "00000000-0000-4000-8000-000000000012",
	action: "staff.login",
	outcome: "failure",
	actor: { kind: "anonymous", id: null, name: null },
	target: { type: "auth_summary", id: null, name: null, organizationId: null, available: false },
};
export const missing: V1ActivityRow = {
	...event,
	id: "00000000-0000-4000-8000-000000000013",
	outcome: "denied",
	actor: { kind: "system", id: null, name: null },
	target: { ...event.target, name: null, available: false },
};
export const list: V1ActivityListBody = {
	items: [event, summary, missing],
	pagination: { page: 1, pageSize: 25, total: 58 },
};
export const customerEvent: V1ActivityRow = {
	...event,
	id: "00000000-0000-4000-8000-000000000014",
	action: "customer.membership.role",
	target: {
		type: "organization_membership",
		id: "00000000-0000-4000-8000-000000000301",
		organizationId: "00000000-0000-4000-8000-000000000101",
		name: "Nadia Prameswari · Koperasi Nusantara",
		available: true,
	},
};
export const customerDetail: V1ActivityDetail = {
	...customerEvent,
	detailAvailable: true,
	changedFields: ["role"],
	previous: null,
	current: null,
};
