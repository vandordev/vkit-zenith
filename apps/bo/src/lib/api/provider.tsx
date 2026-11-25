import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { publicBOConfigSchema } from "@repo/config/public";
import { getBOPublicConfig } from "../../server/public-config";
import { createBOTransport, type BOTransport } from "./client";

type APIState =
	{ status: "pending" } | { status: "error" } | { status: "ready"; transport: BOTransport; epoch: number; invalidated: boolean };
const Context = createContext<APIState>({ status: "pending" });
export function BOProvider({ children }: { children: ReactNode }) {
	const [queries] = useState(
		() =>
			new QueryClient({
				defaultOptions: {
					queries: { retry: false, staleTime: 0, refetchOnWindowFocus: false, refetchOnReconnect: false },
					mutations: { retry: false, gcTime: 0 },
				},
			}),
	);
	const [state, setState] = useState<APIState>({ status: "pending" });
	useEffect(() => {
		let active = true;
		void getBOPublicConfig()
			.then((value) => {
				const config = publicBOConfigSchema.parse(value);
				const transport = createBOTransport(config, () => {
					void queries.cancelQueries();
					queries.clear();
					if (active)
						setState((current) => (current.status === "ready" ? { ...current, epoch: current.epoch + 1, invalidated: true } : current));
				});
				if (active) setState({ status: "ready", transport, epoch: 0, invalidated: false });
			})
			.catch(() => {
				if (active) setState({ status: "error" });
			});
		return () => {
			active = false;
			void queries.cancelQueries();
			queries.clear();
		};
	}, [queries]);
	return (
		<QueryClientProvider client={queries}>
			<Context.Provider value={state}>{children}</Context.Provider>
		</QueryClientProvider>
	);
}
export const useBO = () => useContext(Context);
