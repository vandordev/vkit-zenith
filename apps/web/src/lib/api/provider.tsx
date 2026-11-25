import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { publicWebConfigSchema, type PublicWebConfig } from "@repo/config/public";
import { RouteError, RoutePending } from "../../app/-components/route-states";
import { getPublicConfig } from "../../server/public-config";
import { createAuthTransport, type AuthTransport } from "./client";

type ApiState = { status: "pending" } | { status: "error" } | { status: "ready"; transport: AuthTransport; config: PublicWebConfig };
const Context = createContext<ApiState>({ status: "pending" });
const EpochContext = createContext(0);
export function AuthApiProvider({ children }: { children: ReactNode }) {
	const [queries] = useState(
		() =>
			new QueryClient({
				defaultOptions: {
					queries: { retry: false, staleTime: 0, refetchOnWindowFocus: false, refetchOnReconnect: false },
					mutations: { retry: false, gcTime: 0 },
				},
			}),
	);
	const [state, setState] = useState<ApiState>({ status: "pending" });
	const [epoch, setEpoch] = useState(0);
	useEffect(() => {
		let active = true;
		void getPublicConfig()
			.then((value) => {
				const config = publicWebConfigSchema.parse(value);
				const transport = createAuthTransport(config, () => {
					void queries.cancelQueries();
					queries.clear();
					setEpoch((value) => value + 1);
				});
				if (active) setState({ status: "ready", transport, config });
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
			<Context.Provider value={state}>
				<EpochContext.Provider value={epoch}>{children}</EpochContext.Provider>
			</Context.Provider>
		</QueryClientProvider>
	);
}
// Query keys must include `transport.epoch()` so identity changes never reuse cached data.
export const useAuthApi = () => useContext(Context);
/** Changes on every identity change; key a subtree by it to remount it (drops component state such as one-time secrets). */
export const useIdentityEpoch = () => useContext(EpochContext);

/** Route content renders only once public config and the transport are ready. */
export function ApiGate({ children }: { children: (api: { transport: AuthTransport; config: PublicWebConfig }) => ReactNode }) {
	const api = useAuthApi();
	if (api.status === "pending") return <RoutePending />;
	if (api.status === "error") return <RouteError />;
	return <>{children(api)}</>;
}
