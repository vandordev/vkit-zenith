import { statusPresentation, type StatusKind } from "../../-components/status-badge";

/** Select options for a status filter: the shared Indonesian labels, preceded by "all". */
export const statusOptions = (statuses: readonly string[], kind: StatusKind) => [
	{ value: "", label: "Semua status" },
	...statuses.map((value) => ({ value, label: statusPresentation(kind, value).label })),
];
