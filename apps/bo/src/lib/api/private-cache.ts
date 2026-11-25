import type { QueryClient } from "@tanstack/react-query";

// Existing BO writes deliberately invalidate every private namespace. This also
// covers customer and activity consequences.
export async function invalidatePrivateQueries(queries: QueryClient) {
	await queries.invalidateQueries({ queryKey: ["private"] });
}

export function isStaffGovernanceQuery(entry: { queryKey: readonly unknown[] }) {
	return entry.queryKey[0] === "private" && ["staff", "staff-detail", "staff-sessions"].includes(String(entry.queryKey[3]));
}
