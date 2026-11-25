import { createContext, useContext, useMemo, useState, type ReactNode } from "react";

export type ShellContext = { kind: "organization" | "user" | "staff"; id: string; title: string };
type Setter = { setContext: (context: ShellContext | null) => void };
const Value = createContext<ShellContext | null>(null);
const Channel = createContext<Setter | undefined>(undefined);

export function ShellContextProvider({ children }: { children: ReactNode }) {
	const [context, setContext] = useState<ShellContext | null>(null);
	const channel = useMemo<Setter>(() => ({ setContext }), []);
	return (
		<Channel.Provider value={channel}>
			<Value.Provider value={context}>{children}</Value.Provider>
		</Channel.Provider>
	);
}
/** Child routes publish the entity the shell drills into; they never render navigation. */
export function useShellContext(): Setter {
	const channel = useContext(Channel);
	if (!channel) throw new Error("ShellContextProvider missing");
	return channel;
}
export const useShellContextValue = () => useContext(Value);
