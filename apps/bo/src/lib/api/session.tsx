import { createContext, useContext, type ReactNode, type MutableRefObject } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useBO } from "./provider";
import { boV1GetMe } from "./generated/sdk.gen";
import { zBoV1GetMeResponse } from "./generated/zod.gen";
import type { BoV1GetMeResponse } from "./generated/types.gen";
import type { BOTransport } from "./client";
import { invalidatePrivateQueries } from "./private-cache";
export function useSession() {
	const api = useBO();
	const query = useQuery({
		queryKey: ["identity", api.status === "ready" ? api.epoch : -1],
		enabled: api.status === "ready" && !api.invalidated,
		retry: false,
		queryFn: async ({ signal }) => {
			if (api.status !== "ready") throw new Error("API unavailable");
			return zBoV1GetMeResponse.parse((await boV1GetMe({ client: api.transport.client, throwOnError: true, signal })).data);
		},
	});
	return { api, query };
}
type SessionContext = { staff: BoV1GetMeResponse; transport: BOTransport; epoch: number; writing: MutableRefObject<boolean> };
const Protected = createContext<SessionContext | undefined>(undefined);
export function SessionProvider({ value, children }: { value: SessionContext; children: ReactNode }) {
	return <Protected.Provider value={value}>{children}</Protected.Provider>;
}
export function useProtected() {
	const value = useContext(Protected);
	if (!value) throw new Error("Protected context missing");
	return value;
}
export function usePrivateQuery<T>(
	key: readonly unknown[],
	read: (transport: BOTransport, signal: AbortSignal) => Promise<T>,
	enabled = true,
) {
	const { staff, transport, epoch } = useProtected();
	return useQuery({
		queryKey: ["private", staff.id, epoch, ...key],
		enabled,
		queryFn: ({ signal }) => read(transport, signal),
		retry: false,
	});
}
export function useInvalidate() {
	const queries = useQueryClient();
	return async () => {
		await invalidatePrivateQueries(queries);
	};
}
